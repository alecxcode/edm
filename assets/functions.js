/* This brings all sums to the format: [-]# ###.## on load */
(function() {
  let caseSumsToSep = document.querySelectorAll('span.sum');
  caseSumsToSep.forEach(makeNumSep);
  function makeNumSep(item, index) {
    let s = item.innerText;
    if (s == '' || s == ' ') {
      item.innerHTML = '';
    } else {
      let rS = "";
      let c = 0;
      for (let i = (s.length - 1); i >= 0; i--) {
        rS += s[i];
        c++;
        if (s[i] == ".") { c = 0 };
        if (c == 3) {
          rS += " ";
          c = 0;
        }
      }
      let finalS = "";
      for (let i = (rS.length - 1); i >= 0; i--) {
        finalS += rS[i];
      }
      item.innerHTML = finalS.replace('- ', '-');
    }
  }
})();

/* This brings all sums to the format: [-]####.## on event */
function normalizeSum(elem) {
  setTimeout(function() {
    s = elem.value.replace(/[^-0-9.,]+/g, "").replace(/[,.]+/g, ".");
    if (s) {
      elem.value = s[0] + s.slice(1).split("-").join("");
    } else {
      elem.value = s;
    }
  }, 200);
}

/* This scrolls anchors to normal position so navbar does not overflow */
(function() {
  window.addEventListener('load', scrollToCorrectPosition, false);
  window.addEventListener('hashchange', scrollToCorrectPosition, false);
  function scrollToCorrectPosition() { 
    const urlAnchorRegExp = /#([0-9A-Za-z]+)$/;
    let match = window.location.href.match(urlAnchorRegExp);
    if (match) {
      window.scrollTo(match[1].text);
      window.scrollBy(0,-44);
    }
  }
})();

/* This function handles sum inputs */
function setSumInput(inputID, currencyCode) {
  const inputCurrencyText = document.getElementById(inputID);
  if (inputCurrencyText) {
    const options = document.getElementById(inputID+'List').children;
    for (let i = 0; i < options.length; i++) {
      if (options[i].getAttribute('data-value') == currencyCode) {
        inputCurrencyText.value = options[i].innerText;
        break;
      }
    }
  }
}

/* Function below manupulates chekboxes of the main table */
function processTableSelection(removeAllowed) {
  
  const checkhead = document.getElementById('chead');
  const checkboxes = document.querySelectorAll('input[class="chbox"]');
  const styleRoot = getComputedStyle(document.body);
  const tableOddRowBkgColor = styleRoot.getPropertyValue('--tableOddRowBkgColor');
  const tableSelectionColor = styleRoot.getPropertyValue('--tableSelectionColor');
  const cbListLength = checkboxes.length;
  const deleteButton = document.getElementById('deleteButton');
  const statusButtons = document.querySelectorAll('input[class="sbut statusMultiControl"]');
  
  for (let i = 0; i < cbListLength; i++) {
    let cb = checkboxes[i];
    cb.onclick = function() {
      let state = cb.checked;
      let currentRow = this.parentNode.parentNode;
      if (state) {
        currentRow.style.backgroundColor = tableSelectionColor;
        if (removeAllowed) deleteButton.disabled = false;
        if (statusButtons) statusButtons.forEach(btn => btn.disabled = false);
      } else {
        processAllCBDisabledCheckup();
        currentRow.style.backgroundColor = "inherit";
        if (currentRow.rowIndex % 2 == 0) {
          currentRow.style.backgroundColor = "inherit";
        } else {
          currentRow.style.backgroundColor = tableOddRowBkgColor;
        }
      }
    }
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
  
  checkhead.onclick = function() {
    let chstate = checkhead.checked;
    if (chstate) {
      checkAll();
      if (removeAllowed) deleteButton.disabled = false;
      if (statusButtons) statusButtons.forEach(btn => btn.disabled = false);
    } else {
      uncheckAll();
      if (removeAllowed) deleteButton.disabled = true;
      if (statusButtons) statusButtons.forEach(btn => btn.disabled = true);
    }
  }
  
  function checkAll() {
    checkboxes.forEach(function(cb) {
      cb.checked = true;
      let currentRow = cb.parentNode.parentNode;
      currentRow.style.backgroundColor = tableSelectionColor;
    });
  }
  
  function uncheckAll() {
    checkboxes.forEach(function(cb) {
      cb.checked = false;
      let currentRow = cb.parentNode.parentNode;
      if (currentRow.rowIndex % 2 == 0) {
        currentRow.style.backgroundColor = "inherit";
      } else {
        currentRow.style.backgroundColor = tableOddRowBkgColor;
      }
    });
  }
  
}

/* Resets filters */
function resetFilter() {
  document.getElementById('controlForm').reset();
  addSortingOnSubmut('controlForm');
  document.getElementById('controlForm').submit();
}

/* Sorting functionality */
function applySortingSelection() {
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

/* Filtering functionality */

/* User profile adding for a user filter */
function addUserProfile(profileID, divID, selectorID) {
  if (!profileID) return;
  let inputName = divID;
  let selectorElem = document.getElementById(selectorID);
  let profileName = getSelectInputTextByValue(selectorElem, profileID);
  let divDisp = document.getElementById(divID);

  let arrOfProfileNames = [...divDisp.children].map(elem => elem.innerText);
  if (arrOfProfileNames.includes(profileName)) return;

  let visibleElem = document.createElement('p');
  visibleElem.className = 'chosen-block'
  visibleElem.innerHTML = profileName;
  divDisp.appendChild(visibleElem);
  addHiddenElem(document.getElementById('controlForm'), inputName, profileID);
}

/* This function handles categories and types filters with Selector */
function applyTypeFilterSelector(arrayOfTypes, divID, selectorID) {
  if (arrayOfTypes) {
    arrayOfTypes.forEach(profileID => addUserProfile(profileID, divID, selectorID));
  }
}

/* This function handles categories and types filters */
function applyTypeFilter(arrayOfTypes, flBoxClass) {
  //console.log(arrayOfTypes);
  if (arrayOfTypes) {
    const filterboxes = document.getElementsByName(flBoxClass);
    filterboxes.forEach(function(cb) {
      cb.checked = (arrayOfTypes.includes(+cb.value)) ? true : false;
    });
  }
}

/* This function handles date filters applied from server */
function applyDateFilter(dateFilter) {
  //dateFilter.DatesStr.map(v => v.replace(' ', 'T'));
  if (dateFilter.DatesStr.length == 1) {
    document.getElementById(dateFilter.Name+'Single').value = dateFilter.DatesStr[0];
    document.getElementById(dateFilter.Name+'Relation').value = dateFilter.Relation;
  }
  if (dateFilter.DatesStr.length == 2) {
    const cb = document.getElementById(dateFilter.Name+'Diapason');
    cb.checked = true;
    handleNumericFilterChkBox(cb, dateFilter.Name);
    handleNumericOption(cb, dateFilter.Name);
    document.getElementById(dateFilter.Name+'Start').value = dateFilter.DatesStr[0];
    document.getElementById(dateFilter.Name+'Finish').value = dateFilter.DatesStr[1];
  }
}

/* This function handles sum filters applied from server */
function applySumFilter(sumFilter) {
  const filterCurrency = document.getElementById(sumFilter.Name+'Currency');
  if (filterCurrency) {
    const options = document.getElementById(sumFilter.Name+'CurrencyCodesList').children;
    for (let i = 0; i < options.length; i++) {
      if (options[i].getAttribute('data-value') == sumFilter.CurrencyCode) {
        filterCurrency.value = options[i].innerText;
        break;
      }
    }
  }
  if (sumFilter.Sums.length == 1) {
    document.getElementById(sumFilter.Name+'Single').value = sumFilter.SumsStr[0];
    document.getElementById(sumFilter.Name+'Relation').value = sumFilter.Relation;
  }
  if (sumFilter.Sums.length == 2) {
    const cb = document.getElementById(sumFilter.Name+'Diapason');
    cb.checked = true;
    handleNumericFilterChkBox(cb, sumFilter.Name);
    handleNumericOption(cb, sumFilter.Name);
    document.getElementById(sumFilter.Name+'Start').value = sumFilter.SumsStr[0];
    document.getElementById(sumFilter.Name+'Finish').value = sumFilter.SumsStr[1];
  }
}

function handleNumericFilterChkBox(cb, elemIDPrefix) {
  if (cb.checked == true) {
    document.getElementById(elemIDPrefix+'Single').disabled = true;
    document.getElementById(elemIDPrefix+'Start').disabled = false;
    document.getElementById(elemIDPrefix+'Finish').disabled = false;
  } else {
    document.getElementById(elemIDPrefix+'Single').disabled = false;
    document.getElementById(elemIDPrefix+'Start').disabled = true;
    document.getElementById(elemIDPrefix+'Finish').disabled = true;
  }
}
function handleNumericOption(cb, elemIDPrefix) {
  if (cb.checked == true) {
    document.getElementById(elemIDPrefix+'Relation').disabled = true;
  } else {
    document.getElementById(elemIDPrefix+'Relation').disabled = false;
  }
}

/* This function handles hidden currency input value on input */
function handleNumericFilterList(sourceInput, hiddenInputID) {
  const hiddenInput = document.getElementById(hiddenInputID);
  if (!sourceInput.value) {
    hiddenInput.value = "0";
    return;
  }
  let list = sourceInput.getAttribute('list');
  let options = document.querySelectorAll('#' + list + ' option');
  for (let i = 0; i < options.length; i++) {
    if (options[i].innerText == sourceInput.value) {
      hiddenInput.value = options[i].getAttribute('data-value');
      break;
    }
  }
}

/* This function handles text search filter applied from server */
function applyTextFilter(searchPhrase) {
  document.getElementById('searchText').value = searchPhrase;
}

/* Iterate over table data cells to insert a highlight tag */
function highlightSearchResults(textFilter, columns) {
  textFilter = textFilter.toLowerCase().replace('<', '&lt;').replace('>', '&gt;');
  let tds;
  tb = document.getElementById('mainTable');
  if (tb) {
    tds = tb.getElementsByTagName('td');
  }
  if (textFilter && tds) {
    for (let td of tds) {
      if (columns.indexOf(td.cellIndex) !== -1) {
        td.innerHTML = insertCaseInsensitive(td.innerHTML.replace('&amp;', '&'), textFilter, '<span class="highlight">', '</span>');
        let subElems = td.children;
        for (let i = 0; i < subElems.length; i++) {
          subElems[i].classList.add('clampbig');
        }
      }
    }
  }
}

/* Insert a highlight tag */
function insertCaseInsensitive(srcStr, lowerCaseFilter, before, after) {
  let lowStr = srcStr.toLowerCase();
  let flen = lowerCaseFilter.length;
  let i = -1;
  while ((i = lowStr.indexOf(lowerCaseFilter, i + 1)) != -1) {
    if (insideTag(i, srcStr)) continue;
    srcStr = srcStr.slice(0, i) + before + srcStr.slice(i, i+flen) + after + srcStr.slice(i+flen);
    lowStr = srcStr.toLowerCase();
    i += before.length + after.length;
  }
  return srcStr;
}

/* Check if an ocurrence is inside any tag by index */
function insideTag(si, s) {
  let ahead = false;
  let back = false;
  for (let i = si; i < s.length; i++) {
    if (s[i] == "<") {
      break;
    }
    if (s[i] == ">") {
      ahead = true;
      break;
    }
  }
  for (let i = si; i >= 0; i--) {
    if (s[i] == ">") {
      break;
    }
    if (s[i] == "<") {
      back = true;
      break;
    }
  }
  return (ahead && back);
}

/* Values adding and removal */
function makeFormAndSubmit(id, elemName, action){
  if (!id) return;
  let frm = document.createElement('form');
  frm.method = 'POST';
  frm.action = action;
  let elem = document.createElement('input');
  elem.type = 'hidden';
  elem.name = 'participantAdd';
  elem.name = elemName
  elem.value = id;
  frm.appendChild(elem);
  document.body.appendChild(frm);
  frm.submit();
}



/* File input and file display controls */
function displaySelectedFiles(fileInput, fileDisplay, fileDisplayMsg,
  exceedQuantityMessage, exceedSizeMessage) {
  let res = "";
  for (let eachFile of fileInput.files) {
    res += eachFile.name + '<br>';
  }
  if (!fileQuantityOK(fileInput)) {
    clearInputFiles(fileInput, fileDisplay, fileDisplayMsg, '');
    fileDisplay.innerHTML = '<span class="msgred">'+exceedQuantityMessage+'</span>';
    return;
  }
  if (fileSizeOK(fileInput)) {
    fileDisplay.innerHTML = res;
    fileDisplayMsg.style.display = 'none'; 
  } else {
    clearInputFiles(fileInput, fileDisplay, fileDisplayMsg, '');
    fileDisplay.innerHTML = '<span class="msgred">'+exceedSizeMessage+'</span>';
  }

}
function clearInputFiles(fileInput, fileDisplay, fileDisplayMsg, clearedMessage) {
  if (fileInput.value) {
    try {
      fileInput.value = '';
    } catch(err) { }
    if (fileInput.value) {
      let frm = document.createElement('form');
      let parentNode = fileInput.parentNode;
      let ref = fileInput.nextSibling;
      frm.appendChild(fileInput);
      frm.reset();
      parentNode.insertBefore(fileInput, ref);
    }
  }
  fileDisplay.innerHTML = '<span class="msgok">'+clearedMessage+'</span>';
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
  const MAX_FILES_IN_FORM = 10; /*100 Mb*/
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



/* This function prints message about applied filters */
function printAppliedFilters(fHeading, nofMessage, diapazonMessage, classFilterArr, dateFilterArr, sumFilterArr, textFilter) {
  let resString = "";
  
  for (let classFilter of classFilterArr) {
    
    if (classFilter.Selector) {
      const printedName = document.getElementById(classFilter.Name+'Name').innerText;
      resString += '<span class="af">' + printedName + "</span> ";
      let namesArr = document.getElementById(classFilter.Name).children;
      for (let i = 1; i < namesArr.length; i++) {
        if (namesArr[i].innerText && namesArr[i].innerText != "\n" && namesArr[i].innerText != " ") {
          resString += '<span class="af">' + namesArr[i].innerText + "</span> ";
        }
      }
      resString += '<br>'
    } else {
      const filterboxes = document.getElementsByName(classFilter.Name);
      filterboxes.forEach(function(cb) {
        if (cb.checked) {
          const filterLabel = document.querySelector(`label[for="${cb.id}"]`);
          resString += '<span class="af">' + filterLabel.innerText + "</span> ";
        }
      });
    }
  }

  for (let dateFilter of dateFilterArr) {
    dateFilter.DatesStr = dateFilter.DatesStr.map(v => v.replace('T', ' '));
    const printedName = document.getElementById(dateFilter.Name+'Name').innerText;
    let rel = getSelectInputText(document.getElementById(dateFilter.Name+'Relation'));
    if (dateFilter.DatesStr.length == 1) {
      resString += `<span class="af">${printedName} ${rel}: ${dateFilter.DatesStr[0]}</span> `;
    }
    if (dateFilter.DatesStr.length == 2) {
      resString += `<span class="af">${printedName} ${diapazonMessage}: ${dateFilter.DatesStr[0]} - ${dateFilter.DatesStr[1]}</span> `;
    }
  }

  for (let sumFilter of sumFilterArr) {
    let currency = document.getElementById(sumFilter.Name+'Currency').value;
    if (sumFilter.CurrencyCode > 0) resString += `<span class="af">${currency}</span> `;
    const printedName = document.getElementById(sumFilter.Name+'Name').innerText;
    let rel = getSelectInputText(document.getElementById(sumFilter.Name+'Relation'));
    if (sumFilter.Sums.length == 1) {
      resString += `<span class="af">${printedName} ${rel}: ${sumFilter.SumsStr[0]}</span> `;
    }
    if (sumFilter.Sums.length == 2) {
      resString += `<span class="af">${printedName} ${diapazonMessage}: ${sumFilter.SumsStr[0]} - ${sumFilter.SumsStr[1]}</span> `;
    }
  }
  
  if (textFilter) {
    textFilter = textFilter.replace('<', '&lt;').replace('>', '&gt;');
    resString += '<span class="af">' + textFilter + "</span> ";
  }
  
  const p = document.getElementById('appliedFilters');
  p.innerHTML = (resString.length > 1) ? fHeading + resString : nofMessage;

}

function getSelectInputText(sel) {
  return sel.options[sel.selectedIndex].text;
}

function getSelectInputTextByValue(sel, val) {
  for (let i = 0; i < sel.length; i++) {
    if (sel.options[i].value == val) {
        return sel.options[i].text;
    }
  }
}

/* Pagination mechanics */
function paginate(val) {
  const paginatorControl = document.getElementById('pageNumber');
  let page = Number(paginatorControl.value);
  let maxv = Number(paginatorControl.getAttribute("max"));
  if (val == 'max') val = maxv;
  if (val == 'min') {page = 1; val = 0};
  page += val;
  if (page < 1) page = 1;
  if (page >= maxv ) page = maxv;
  paginatorControl.value = page;
  const frm = 'paginationForm';
  addFiltersOnSubmut(frm);
  addSortingOnSubmut(frm);
  document.getElementById('paginationForm').submit();
}

function calcTotalPages(elemsOnPage, filteredNum) {
  let result = 1;
  let q = Math.floor(filteredNum / elemsOnPage);
  let r = filteredNum % elemsOnPage;
  if (q >= 1) {
    result = q;
    if (r > 0) {
      result += 1;
    }
  }
  return result;
}

function addHiddenElem(theForm, name, value) {
  let input = document.createElement('input');
  input.type = 'hidden';
  input.name = name;
  input.value = value;
  theForm.appendChild(input);
}

function processPagesCalculations(textTotal, textOnPage, textPages, elemsOnPage, filteredNum) {
  let totalPages = calcTotalPages(elemsOnPage, filteredNum);
  document.getElementById('pageNumber').setAttribute('max', totalPages)
  document.querySelector('label[for="pageNumber"]').innerHTML += String(totalPages) + '.';
  let rowCount = getElemsOnCurrentPage();
  document.getElementById('stat').innerHTML = `<br>${textTotal}: ${filteredNum}<br>${textOnPage}: ${rowCount}<br>${textPages}: ${totalPages}`
}

function getElemsOnCurrentPage() {
  return document.getElementById('mainTable').getElementsByTagName('tr').length - 1;
}

/* Text formatting in multiline elements */
function insertBBCode(tag, textInputID) {
  const elem = document.getElementById(textInputID);
  let val = elem.value;
  let selectedText = val.substring(elem.selectionStart, elem.selectionEnd);
  let beforeText = val.substring(0, elem.selectionStart);
  let afterText = val.substring(elem.selectionEnd, val.length);
  elem.value = beforeText + '[' + tag + ']' + selectedText + '[/' + tag + ']' + afterText;
}

function updateMultilines(className) {
  const urlREGex = /((?:http|https|ftp|sftp):\/\/[^<\s\n\r\t]+[^.,;<\s\n\r\t"'])/gim;
  const wwwREGex = /(www\.[^<\s\n\r\t]+[^.,;<\s\n\r\t"'])/gim; 
  let multiLineLabels = document.getElementsByClassName(className);
  if (multiLineLabels) {
    
    for (let elem of multiLineLabels) {
      /*.replace(/\n/g, "<br>")*/
      elem.innerHTML = elem.innerHTML
      .replace(/\[b\]/g, '<b>').replace(/\[\/b\]/g, '</b>')
      .replace(/\[i\]/g, '<i>').replace(/\[\/i\]/g, '</i>')
      .replace(/\[u\]/g, '<u>').replace(/\[\/u\]/g, '</u>')
      .replace(/\[code\]/g, '<pre>').replace(/\[\/code\]/g, '</pre>');
      elem.innerHTML = replaceNewlinesWithBR(elem.innerHTML);
      if (urlREGex.test(elem.innerHTML)) {
        elem.innerHTML = elem.innerHTML.replace(urlREGex, '<a href="$1" target="_blank">$1</a>');
      }
      if (wwwREGex.test(elem.innerHTML)) {
        elem.innerHTML = elem.innerHTML.replace(wwwREGex, '<a href="http://$1" target="_blank">$1</a>');
      }
    }
  }
}

function replaceNewlinesWithBR(cont) {
  let arr = [' ', ' ', ' ', ' ', ' ', ' ', ' '];
	let insidetagname = false;
	let outsidecode = true;
  let arrcounter = 0;
	for (let i = 0; i < cont.length; i++) {
		if (cont[i] == '<') {
			arr[0] = '<';
			insidetagname = true;
      arrcounter = 0;
		}
		if (insidetagname) {
			arr[arrcounter] = cont[i];
      arrcounter++;
      if (arrcounter > 6) arrcounter = 1;
			if (cont[i] == '>') {
        insidetagname = false;
        arrcounter = 0;
      }
		}
		if (arr.join('') == "<code> " || arr.join('') == "<pre>  ") outsidecode = false;
		if (arr.join('') == "</code>" || arr.join('') == "</pre> ") outsidecode = true;
		if (!insidetagname) arr = arr.map(() => ' ');
		if (outsidecode && cont[i] == '\r') cont = cont.slice(0,i) + cont.slice(i+1);
		if (outsidecode && cont[i] == '\n') {
			cont = cont.slice(0,i) + "<br>" +  cont.slice(i+1);
			i += 3;
		}
	}
	return cont;
}

function clearBBCode(queryName) {
  let contentElems = document.querySelectorAll(queryName);
  contentElems.forEach(elem => {
    elem.innerHTML = elem.innerHTML
    .replace(/\[b\]/g, '').replace(/\[\/b\]/g, '')
    .replace(/\[i\]/g, '').replace(/\[\/i\]/g, '')
    .replace(/\[u\]/g, '').replace(/\[\/u\]/g, '')
    .replace(/\[code\]/g, '').replace(/\[\/code\]/g, '');
  });
}

/* ======== Mobile elements ======== */

/* This function shows or hides mobile menu */
function showMobileMenu() {
  const d = document.getElementById('mobilemenu');
  const styleRoot = getComputedStyle(document.body);
  if (d.style.display == 'none') {
    document.getElementById('container').style.display = 'none';
    d.style.display = 'block';
    document.body.style.backgroundColor = styleRoot.getPropertyValue('--menuBackgroundColor');
    d.innerHTML = document.getElementById('textmenu').innerHTML;
    /* if (document.getElementById('showhide').style.display != 'none') {
      showFiltersMobile(document.getElementById('mobileMenuButtonFilters'), 'Close', 'Filters');
    } */
  } else {
    document.getElementById('container').style.display = 'block';
    d.style.display = 'none';
    document.body.style.backgroundColor = styleRoot.getPropertyValue('--defaultBackgroundColor');
    d.innerHTML = '';
  }
  
}
