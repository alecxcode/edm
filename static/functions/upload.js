/* File input and file display controls */
function displaySelectedFiles(fileInput, fileDisplay, fileDisplayMsg,
  exceedQuantityMessage, exceedSizeMessage) {
  let res = "";
  for (let eachFile of fileInput.files) {
    res += eachFile.name + '<br>';
  }
  if (!fileQuantityOK(fileInput)) {
    clearInputFiles(fileInput, fileDisplay, fileDisplayMsg, '');
    fileDisplay.innerHTML = '<span class="msgred">' + exceedQuantityMessage + '</span>';
    return;
  }
  if (fileSizeOK(fileInput)) {
    fileDisplay.innerHTML = res;
    fileDisplayMsg.style.display = 'none';
  } else {
    clearInputFiles(fileInput, fileDisplay, fileDisplayMsg, '');
    fileDisplay.innerHTML = '<span class="msgred">' + exceedSizeMessage + '</span>';
    return;
  }
  const parentNode = fileInput.parentNode;
  const fiid = fileInput.id;
  if (fileInput.value) {
    const oldtmp = document.getElementById('tmp'+fiid);
    if (oldtmp) parentNode.removeChild(oldtmp);
    const tmp = fileInput.cloneNode(true);
    tmp.id = 'tmp'+fiid;
    tmp.disabled = true;
    tmp.style.display = 'none';
    parentNode.insertBefore(tmp, fileInput.nextSibling);
  } else {
    parentNode.removeChild(fileInput);
    const tmp = document.getElementById('tmp'+fiid);
    if (!tmp || !tmp.value) return;
    tmp.id = fiid;
    tmp.disabled = false;
    tmp.style.display = 'unset';
    displaySelectedFiles(tmp, fileDisplay, fileDisplayMsg,
    exceedQuantityMessage, exceedSizeMessage);
  }
}

function clearInputFiles(fileInput, fileDisplay, fileDisplayMsg, clearedMessage) {
  if (fileInput.value) {
    try {
      fileInput.value = '';
    } catch (err) { }
    if (fileInput.value) {
      let frm = document.createElement('form');
      let parentNode = fileInput.parentNode;
      let ref = fileInput.nextSibling;
      frm.appendChild(fileInput);
      frm.reset();
      parentNode.insertBefore(fileInput, ref);
    }
  }
  fileDisplay.innerHTML = '<span class="msgok">' + clearedMessage + '</span>';
  fileDisplayMsg.style.display = 'block';
}

function fileSizeOK(fileInput) {
  const MAX_UPLOAD_SIZE = 104857600; /*100 Mb*/
  let totalsize = 0;
  for (let eachFile of fileInput.files) {
    totalsize += eachFile.size;
  }
  if (totalsize > MAX_UPLOAD_SIZE) {
    return false;
  } else {
    return true;
  }
}

function fileQuantityOK(fileInput) {
  const MAX_FILES_IN_FORM = 100;
  if (fileInput.files.length > MAX_FILES_IN_FORM) {
    return false;
  } else {
    return true;
  }
}

/* File inputs adding on submit for removal */
function addFileElems(frm) {
  const checkboxes = document.querySelectorAll('input[class="fchbox"]');
  for (let i = 0; i < checkboxes.length; i++) {
    if (checkboxes[i].checked) {
      addHiddenElem(frm, checkboxes[i].name, checkboxes[i].value);
    }
  }
}
