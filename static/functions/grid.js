/* Function below processes chekboxes of the main grid */
function processTableSelection(removeAllowed) {

  const checkhead = document.getElementById('chead');
  const checkboxes = document.querySelectorAll('input[class="chbox"]');
  const styleRoot = getComputedStyle(document.body);
  const tableEvenRowBkgColor = styleRoot.getPropertyValue('--defaultBackgroundColor');
  const tableOddRowBkgColor = styleRoot.getPropertyValue('--tableOddRowBkgColor');
  const tableSelectionColor = styleRoot.getPropertyValue('--tableSelectionColor');
  const cbListLength = checkboxes.length;
  const deleteButton = document.getElementById('deleteButton');
  const statusButtons = document.querySelectorAll('input[class="sbut statusMultiControl"]');

  document.getElementById("mainTable").addEventListener("click", function (event) {
    if (event.target.classList.contains('grideven') || event.target.classList.contains('gridodd')) {
      let cb = document.getElementById(+event.target.classList[0].split('-')[1]);
      cb.checked = !cb.checked;
      checkOne(cb.checked, '.' + cb.parentNode.classList[0]);
    }
    if ((event.target.classList.contains('content') || event.target.classList.contains('description')) &&
      (event.target.parentElement.classList.contains('grideven') || (event.target.parentElement.classList.contains('gridodd')))) {
      let cb = document.getElementById(+event.target.parentElement.classList[0].split('-')[1]);
      cb.checked = !cb.checked;
      checkOne(cb.checked, '.' + cb.parentNode.classList[0]);
    }
  });

  for (let i = 0; i < cbListLength; i++) {
    let cb = checkboxes[i];
    cb.onclick = function () {
      checkOne(this.checked, '.' + this.parentNode.classList[0]);
    };
  }

  function checkOne(state, currentRowClass) {
    if (state) {
      changeRowBkgColor(currentRowClass, true);
      if (removeAllowed) deleteButton.disabled = false;
      if (statusButtons) statusButtons.forEach(btn => btn.disabled = false);
    } else {
      processAllCBDisabledCheckup();
      changeRowBkgColor(currentRowClass, false);
    }
  }

  function changeRowBkgColor(classSelector, select) {
    document.querySelectorAll(classSelector).forEach((elem) => {
      if (select) {
        elem.style.backgroundColor = tableSelectionColor;
      } else {
        elem.style.backgroundColor = (elem.classList.contains('gridodd')) ? tableOddRowBkgColor : tableEvenRowBkgColor;
      }
    });
  }

  function processAllCBDisabledCheckup() {
    for (let i = 0; i < cbListLength; i++) {
      if (checkboxes[i].checked) {
        return;
      }
    }
    checkhead.checked = false;
    if (removeAllowed) deleteButton.disabled = true;
    if (statusButtons) statusButtons.forEach(btn => btn.disabled = true);
  }

  checkhead.onclick = function () {
    if (checkhead.checked) {
      checkAll(true);
    } else {
      checkAll(false);
    }
  }

  function checkAll(select) {
    if (removeAllowed) deleteButton.disabled = !select;
    if (statusButtons) statusButtons.forEach(btn => btn.disabled = !select);
    checkboxes.forEach(function (cb) {
      cb.checked = select;
      let currentRowClass = '.' + cb.parentNode.classList[0];
      changeRowBkgColor(currentRowClass, select);
    });
  }

}

/* Hides control buttons and disables checkboxes */
function disableControlButtons(mainTablePresent) {
  document.getElementById('controlButtons').style.display = 'none';
  if (mainTablePresent) {
    const checkhead = document.getElementById('chead');
    const checkboxes = document.querySelectorAll('input[class="chbox"]');
    checkhead.disabled = true;
    checkboxes.forEach(cb => cb.disabled = true);
  }
}

/* Sorting functionality */
function applySortingSelection() {
  if (!document.getElementById('sortedBy') || !document.getElementById('sortedHow')) return;
  let valSortedBy = document.getElementById('sortedBy').value;
  let currentSortedBy = document.getElementById(valSortedBy);
  let nodeList = currentSortedBy.children;
  if (nodeList.length > 2) nodeList = [nodeList[1], nodeList[2]];
  let valSortedHow = document.getElementById('sortedHow').value;
  if (valSortedHow == 0) { valSortedHow = 1; } else if (valSortedHow == 1) { valSortedHow = 0; }
  nodeList[valSortedHow].className = "sortchosen";
}

function sortBy(bywhat, how) {
  let frm = document.getElementById('sortingForm');
  document.getElementById('sortedBy').value = bywhat;
  document.getElementById('sortedHow').value = how;
  addFiltersOnSubmut('sortingForm');
  frm.submit();
}

/* Submitting data */
function submitControlButton(inputName, inputValue) {
  addSortingOnSubmut('datagridForm');
  addFiltersOnSubmut('datagridForm');
  addPaginationOnSubmut('datagridForm');
  const frm = document.getElementById('datagridForm');
  removeEmptyInputs(frm);
  sessionStorage.setItem('lastQuery', location.search);
  addHiddenElem(frm, inputName, inputValue);
  frm.submit();
}
function addSortingOnSubmut(formID) {
  const frm = document.getElementById(formID);
  let sortedBy = document.getElementById('sortedBy');
  let sortedHow = document.getElementById('sortedHow');
  if (frm && sortedBy && sortedHow) {
    addHiddenElem(frm, 'sortedBy', sortedBy.value);
    addHiddenElem(frm, 'sortedHow', sortedHow.value);
  }
}

/* Calculates the number of rows in the main grid */
function getElemsOnCurrentPage() {
  return document.getElementById('mainTable').getElementsByClassName('firstcell').length;
}
