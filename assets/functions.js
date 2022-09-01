/* Highlights current website block */
(function(){
  let addr = window.location.pathname;
  let currentItem;
  if (addr.includes("/docs")) {
    currentItem = document.querySelector("#textmenu a[href^='/docs']");
  } else if (addr.includes("/team")) {
    currentItem = document.querySelector("#textmenu a[href^='/team']");
  } else if (addr.includes("/task")) {
    currentItem = document.querySelector("#textmenu a[href^='/task']");
  }
  if (currentItem) currentItem.classList.add("chosenmenu");
})();

/* Translates the page - aka i18n */
(function(){
  const langCode = document.documentElement.lang;
  if (langCode != 'en') {
    getLang(langCode).then(lang => {
      translateAll(null, lang);
      document.title = lang.appTitle + ': ' + document.querySelector('h1').innerText;
      const loadingDiv = document.querySelector('.fullscreen');
      if (loadingDiv) loadingDiv.parentNode.removeChild(loadingDiv);
    });
  }
})();

/* Translates specific element */
function translateElem(rootElem) {
  const langCode = document.documentElement.lang;
  if (langCode != 'en') {
    getLang(langCode).then(lang => translateAll(rootElem, lang));
  }
}

async function getLang(langCode) {
  const request = new Request(`/assets/i18n/${langCode}.json`);
  const response = await fetch(request, {method: 'GET', credentials: 'include', mode: 'no-cors'});
  if (response.ok) {
    const lang = await response.json();
    return lang;
  } else {
    return {};
  }
}

function translateAll(rootElem, lang) {
  if (!rootElem) {
    rootElem = document;
  }

  function getAdditionalSymbols(s, e) {
    if (!e) return s;
    const c0 = e.charAt(0);
    const c1 = e.charAt(e.length - 1);
    const c2 = e.charAt(e.length - 2);
    if (c0 == '+' && s.charAt(0) != '+') s = '+ ' + s;
    if (c0 == '×' && s.charAt(0) != '×') s = '× ' + s;
    if (c1 == ':' && c2 != '*' && s.charAt(s.length - 1) != ':') {s += ':';
    } else if (c1 == ':' && c2 == '*' && s.charAt(s.length - 2) != '*') {s += '*:';
    } else if (c1 == ' ' && c2 == ':') {s += ': ';}
    return s;
  }

  const elems = rootElem.querySelectorAll('[i18n-text]');
  elems.forEach(elem => {
    let s = lang[elem.getAttribute('i18n-text')];
    if (s) elem.innerHTML = getAdditionalSymbols(s, elem.innerText);
  });

  const alts = rootElem.querySelectorAll('[i18n-alt]');
  alts.forEach(elem => {
    let s = lang[elem.getAttribute('i18n-alt')];
    if (s) elem.alt = s;
  });

  const titles = rootElem.querySelectorAll('[i18n-title]');
  titles.forEach(elem => {
    let s = lang[elem.getAttribute('i18n-title')];
    if (s) elem.title = s;
  });

  const vals = rootElem.querySelectorAll('[i18n-value]');
  vals.forEach(elem => {
    let s = lang[elem.getAttribute('i18n-value')];
    if (s) elem.value = getAdditionalSymbols(s, elem.value);
  });

  const placeholders = rootElem.querySelectorAll('[i18n-placeholder]');
  placeholders.forEach(elem => {
    let s = lang[elem.getAttribute('i18n-placeholder')];
    if (s) elem.placeholder = s;
  });

  const labels = rootElem.querySelectorAll('[i18n-label]');
  labels.forEach(elem => {
    let s = lang[elem.getAttribute('i18n-label')];
    if (s) elem.setAttribute('label', s);
  });

  const many = rootElem.querySelectorAll('[i18n-many]');
  many.forEach(elem => {
    let s = lang[elem.getAttribute('i18n-many')];
    if (s) {
      for (let key of Object.keys(s)) {
        if (key == 'text') {
          elem.innerHTML = getAdditionalSymbols(s[key], elem.innerText);
        } else if (key == 'value') {
          elem.value = getAdditionalSymbols(s[key], elem.value);
        } else {
          elem.setAttribute(key, s[key]);
        }
      }
    }
  });

  const indexes = rootElem.querySelectorAll('[i18n-index]');
  indexes.forEach(elem => {
    let selector = elem.getAttribute('i18n-index').split('-');
    if (lang[selector[0]]) {
      let s = lang[selector[0]][+selector[1]];
      if (s) elem.innerHTML = s;
    }
  });

  const sorters = rootElem.querySelectorAll('span.sort, span.sortchosen');
  sorters.forEach(elem => {
    if (lang.ascending && lang.descending) {
      let t = elem.getAttribute('title');
      if (t == 'Ascending') {
        elem.title = lang.ascending;
      } else if (t == 'Descending') {
        elem.title = lang.descending;
      }
    }
  });

}

/* Replaces a value of an agrument in an URL get request */
function replaceGetArg(url, key, val){
  let res = '?';
  var arr = url.split('?')[1].split('&');
  if (arr) {
    for (let i = 0; i < arr.length; i++){
      if (arr[i].split('=')[0] == key){
        res += key + '=' + val + '&';
      } else {
        res += arr[i] + '&'
      }
    }
  }
  return res.slice(0, -1);
}
/* Replaces the page number arg in the URL according to last query after POST */
function replacePageNumberArg(pageNumber){
  let lastQuery = sessionStorage.getItem('lastQuery');
  if (lastQuery) {
    sessionStorage.removeItem('lastQuery');
    lastQuery = replaceGetArg(lastQuery, 'pageNumber', pageNumber);
    window.history.replaceState(null, '', lastQuery);
  }
}

/* Check for newly created object */
function checkForNewCreated() {
  let result = sessionStorage.getItem('new');
  if (result) {
    sessionStorage.removeItem('new');
    if (!window.location.href.endsWith('new')) {
      let msgnew = document.createElement('div');
      msgnew.className = "msgok";
      msgnew.innerHTML = "Creation completed.";
      const langCode = document.documentElement.lang;
      if (langCode != 'en') {
        getLang(langCode).then(lang => {if (lang.creationCompleted) msgnew.innerHTML = lang.creationCompleted});
      }
      document.getElementById('control').appendChild(msgnew);
    }
  }
}

/* This function shows or hides edit controls */
function showEditForm(what) {
  const view = document.getElementById('view');
  const edit = document.getElementById('edit');
  const removal = document.getElementById('removal');
  const removalFiles = document.getElementById('removalFiles');
  if (what == 'edit') {
    view.style.display = 'none';
    edit.style.display = 'block';
    removal.style.display = 'none';
    if (removalFiles) removalFiles.style.display = 'none'
  } else if (what == 'removal') {
    view.style.display = 'block';
    edit.style.display = 'none';
    removal.style.display = 'block';
    if (removalFiles) removalFiles.style.display = 'none'
  } else if (what == 'removalFiles') {
    view.style.display = 'block';
    edit.style.display = 'none';
    removal.style.display = 'none';
    if (removalFiles) removalFiles.style.display = 'block'
  } else {
    view.style.display = 'block';
    edit.style.display = 'none';
    removal.style.display = 'none';
    if (removalFiles) removalFiles.style.display = 'none'
  }
}

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
    let s = elem.value.replace(/[^-0-9.,]+/g, "").replace(/[,.]+/g, ".");
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

/* This removes unused inputs from form submission */
function removeEmptyInputs(frm) {
  const inputs = frm.getElementsByTagName('input');
  for (i = 0; i < inputs.length; i++) {
    if (!inputs[i].value) inputs[i].disabled = true;
  }
  const selects = frm.getElementsByTagName('select');
  for (i = 0; i < selects.length; i++) {
    if (!selects[i].value) selects[i].disabled = true;
    if (selects[i].value == 'eq') {
      let numInpID = selects[i].name.replace('Relation', 'Single');
      if (!document.getElementById(numInpID).value) selects[i].disabled = true;
    }
  }
}

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
  const tableEvenRowBkgColor = styleRoot.getPropertyValue('--defaultBackgroundColor');
  const tableOddRowBkgColor = styleRoot.getPropertyValue('--tableOddRowBkgColor');
  const tableSelectionColor = styleRoot.getPropertyValue('--tableSelectionColor');
  const cbListLength = checkboxes.length;
  const deleteButton = document.getElementById('deleteButton');
  const statusButtons = document.querySelectorAll('input[class="sbut statusMultiControl"]');

  document.getElementById("mainTable").addEventListener("click", function (event) {
    if (event.target.classList.contains('grideven') || event.target.classList.contains('gridodd')) {
      let id = (+event.target.classList[0].split('-')[1]);
      let cb = document.getElementById(id);
      cb.checked = !cb.checked;
      checkOne(cb.checked, '.' + cb.parentNode.classList[0]);
    }
  });
  
  for (let i = 0; i < cbListLength; i++) {
    let cb = checkboxes[i];
    cb.onclick = function() {
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

  function changeRowBkgColor(classSelector, select){
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
  
  checkhead.onclick = function() {
    if (checkhead.checked) {
      checkAll(true);
    } else {
      checkAll(false);
    }
  }
  
  function checkAll(select) {
    if (removeAllowed) deleteButton.disabled = !select;
    if (statusButtons) statusButtons.forEach(btn => btn.disabled = !select);
    checkboxes.forEach(function(cb) {
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

/* Resets filters */
function resetFilter() {
  const frm = document.getElementById('controlForm');
  frm.reset();
  removeEmptyInputs(frm);
  addSortingOnSubmut('controlForm');
  frm.submit();
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

/* This function sets inputs accroding to applied filters */
function applyAllFilters(filters, classFilterArr, dateFilterArr, sumFilterArr) {
  let classFilter = filters.ClassFilter;
  let classFilterOR = filters.ClassFilterOR;
  let dateFilter = filters.DateFilter;
  let sumFilter = filters.SumFilter;
  let textFilter = filters.TextFilter;

  if (classFilter) classFilter.forEach((e) => classFilterArr.push(e));
  if (classFilterOR) classFilterOR.forEach((e) => {
    if (!classFilterArr.some(v => v.Name == e.Name)) {
      classFilterArr.push(e);
    }
  });

  if (classFilterArr) {
    for (let eachF of classFilterArr) {
      if (eachF.Selector) {
        applyTypeFilterSelector(eachF.List, eachF.Name, eachF.Selector);
      } else {
        applyTypeFilter(eachF.List, eachF.Name);
      }
    }
  }

  if (dateFilter) {
    for (let eachF of dateFilter) {
      applyDateFilter(eachF);
      dateFilterArr.push(eachF);
    }
  }

  if (sumFilter) {
    for (let eachF of sumFilter) {
      applySumFilter(eachF);
      sumFilterArr.push(eachF);
    }
  }
  
  if (textFilter) {
    applyTextFilter(textFilter);
  }
}

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
  const filterCurrencyCode = document.getElementById(sumFilter.Name+'CurrencyCode');
  if (filterCurrency) {
    const options = document.getElementById(sumFilter.Name+'CurrencyCodesList').children;
    for (let i = 0; i < options.length; i++) {
      if (options[i].getAttribute('data-value') == sumFilter.CurrencyCode) {
        filterCurrency.value = options[i].innerText;
        filterCurrencyCode.value = options[i].getAttribute('data-value');
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
function highlightSearchResults(textFilter) {
  textFilter = textFilter.toLowerCase().replace('<', '&lt;').replace('>', '&gt;');
  let elems;
  const content = document.getElementById('mainTable');
  if (content) {
    elems = content.getElementsByClassName('textsearch');
  }
  if (textFilter && elems) {
    for (let elem of elems) {
      recursiveChildrenSearch(elem, textFilter);
    }
  }
}

function recursiveChildrenSearch(elem, textFilter) {
  if (elem.children && elem.children.length > 0) {
    for (subelem of elem.children) {
      recursiveChildrenSearch(subelem, textFilter);
    }
  } else {
    if (elem.classList && elem.classList.contains('mobile')) return;
    elem.innerHTML = insertCaseInsensitive(elem.innerHTML.replace('&amp;', '&'), textFilter, '<span class="highlight">', '</span>');
  }
}

/* Insert a highlight tag */
function insertCaseInsensitive(srcStr, lowerCaseFilter, before, after) {
  let lowStr = srcStr.toLowerCase();
  let flen = lowerCaseFilter.length;
  let i = -1;
  while ((i = lowStr.indexOf(lowerCaseFilter, i + 1)) != -1) {
    //if (insideTag(i, srcStr)) continue;
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
  frm.method = 'post';
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
  addHiddenElem(frm, 'sortedBy', document.getElementById('sortedBy').value);
  addHiddenElem(frm, 'sortedHow', document.getElementById('sortedHow').value);
}
function addStandardPagination(formID) {
  const frm = document.getElementById(formID);
  let pageNumber = document.getElementById('pageNumber').value;
  let elemsOnPage = document.getElementById('elemsOnPage').value;
  let elemsOnCurrentPage = getElemsOnCurrentPage();
  addHiddenElem(frm, 'elemsOnCurrentPage', elemsOnCurrentPage);
  addHiddenElem(frm, 'pageNumber', pageNumber);
  addHiddenElem(frm, 'elemsOnPage', elemsOnPage);
}
function addSeekPagination(formID, filteredNum){
  const frm = document.getElementById(formID);
  let pageNumber = document.getElementById('pageNumber').value;
  let elemsOnPage = document.getElementById('elemsOnPage').value;
  let elemsOnCurrentPage = getElemsOnCurrentPage();
  addHiddenElem(frm, 'elemsOnCurrentPage', elemsOnCurrentPage);
  addHiddenElem(frm, 'pageNumber', pageNumber);
  addHiddenElem(frm, 'elemsOnPage', elemsOnPage);
  addHiddenElem(frm, 'filteredNum', filteredNum);
  addHiddenElem(frm, 'previousPage', pageNumber);
  const mainTable = document.getElementById('mainTable');
  //console.log(mainTable.getElementsByClassName('firstcell')[0].lastElementChild.innerText, mainTable.getElementsByClassName('firstcell')[mainTable.getElementsByClassName('firstcell').length-1].lastElementChild.innerText);  
  addHiddenElem(frm, 'firstElemOnPage', mainTable.getElementsByClassName('firstcell')[0].lastElementChild.innerText);
  addHiddenElem(frm, 'lastElemOnPage',  mainTable.getElementsByClassName('firstcell')[mainTable.getElementsByClassName('firstcell').length-1].lastElementChild.innerText);
}
function processAddingFilters(frm, filters) {

  let classFilter = filters.ClassFilter;
  let classFilterOR = filters.ClassFilterOR;
  let dateFilter = filters.DateFilter;
  let sumFilter = filters.SumFilter;
  let textFilter = filters.TextFilter;

  if (classFilter) {
    for (let eachF of classFilter) {
      for (let i = 0; i < eachF.List.length; i++) {
        addHiddenElem(frm, eachF.Name, eachF.List[i]);
      }
    }
  }

  if (classFilterOR) {
    let currentCFORName;
    for (let eachF of classFilterOR) {
      if (eachF.Name != currentCFORName) {
        currentCFORName = eachF.Name;
        for (let i = 0; i < eachF.List.length; i++) {
          addHiddenElem(frm, eachF.Name, eachF.List[i]);
        }
      }
    }
  }

  if (dateFilter) {
    for (let eachF of dateFilter) {
      addHiddenElem(frm, eachF.Name+'Relation', eachF.Relation);
      for (let i = 0; i < eachF.Dates.length; i++) {
        addHiddenElem(frm, eachF.Name, eachF.DatesStr[i]);
      }
    }
  }

  if (sumFilter) {
    for (let eachF of sumFilter) {
      addHiddenElem(frm, eachF.Name+'CurrencyCode', eachF.CurrencyCode);
      addHiddenElem(frm, eachF.Name+'Relation', eachF.Relation);
      for (let i = 0; i < eachF.SumsStr.length; i++) {
        addHiddenElem(frm, eachF.Name, eachF.SumsStr[i]);
      }
    }
  }

  if (textFilter) {
    addHiddenElem(frm, 'searchText', textFilter);
  }
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

/* This function prints message about applied filters */
function printAppliedFilters(classFilterArr, dateFilterArr, sumFilterArr, textFilter) {

  let resString = "";
  let timeout = 0;
  var runAllowed = false;
  var retries = 0;

  let appliedFilters = "Applied filters: ";
  let noFiltersApplied = "No filters applied.";
  let diapason = "Interval";
  const langCode = document.documentElement.lang;
  if (langCode != 'en') {
    timeout = 20;
    getLang(langCode).then(lang => {
      if (lang.appliedFilters) appliedFilters = lang.appliedFilters + ": ";
      if (lang.noFiltersApplied) noFiltersApplied = lang.noFiltersApplied;
      if (lang.diapason) diapason = lang.diapason;
      runAllowed = true;
    });
  } else {
    runAllowed = true;
  }
  
  var intervalID = setInterval(function() {
    retries++
    if (retries > 10000) clearInterval(intervalID);
    if (!document.querySelector('.fullscreen') && runAllowed) {
      clearInterval(intervalID);
      for (let classFilter of classFilterArr) {
        if (classFilter.Selector) {
          const printedName = document.getElementById(classFilter.Name+'Display').innerText;
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
        const printedName = document.getElementById(dateFilter.Name+'Display').innerText;
        let rel = getSelectInputText(document.getElementById(dateFilter.Name+'Relation'));
        if (dateFilter.DatesStr.length == 1) {
          resString += `<span class="af">${printedName} ${rel}: ${dateFilter.DatesStr[0]}</span> `;
        }
        if (dateFilter.DatesStr.length == 2) {
          resString += `<span class="af">${printedName} ${diapason}: ${dateFilter.DatesStr[0]} - ${dateFilter.DatesStr[1]}</span> `;
        }
      }
      for (let sumFilter of sumFilterArr) {
        let currency = document.getElementById(sumFilter.Name+'Currency').value;
        if (sumFilter.CurrencyCode > 0) resString += `<span class="af">${currency}</span> `;
        const printedName = document.getElementById(sumFilter.Name+'Display').innerText;
        let rel = getSelectInputText(document.getElementById(sumFilter.Name+'Relation'));
        if (sumFilter.Sums.length == 1) {
          resString += `<span class="af">${printedName} ${rel}: ${sumFilter.SumsStr[0]}</span> `;
        }
        if (sumFilter.Sums.length == 2) {
          resString += `<span class="af">${printedName} ${diapason}: ${sumFilter.SumsStr[0]} - ${sumFilter.SumsStr[1]}</span> `;
        }
      }
      if (textFilter) {
        textFilter = textFilter.replace('<', '&lt;').replace('>', '&gt;');
        resString += '<span class="af">' + textFilter + "</span> ";
      }
      const p = document.getElementById('appliedFilters');
      p.innerHTML = (resString.length > 1) ? appliedFilters + resString : noFiltersApplied;
    }
  }, timeout);
  
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
  removeEmptyInputs(document.getElementById(frm));
  document.getElementById(frm).submit();
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

function processPagesCalculations(elemsOnPage, filteredNum) {
  let totalPagesNum = calcTotalPages(elemsOnPage, filteredNum);
  document.getElementById('pageNumber').setAttribute('max', totalPagesNum)
  document.getElementById('totalPagesNumber').innerHTML = String(totalPagesNum);
  let rowCount = getElemsOnCurrentPage();
  let totalElemsFound = "Total objects by search criteria";
  let onThisPage = "On this page";
  let totalPages = "Total pages";
  const langCode = document.documentElement.lang;
  if (langCode != 'en') {
    getLang(langCode).then(lang => {
      if (lang.totalElemsFound) totalElemsFound = lang.totalElemsFound;
      if (lang.onThisPage) onThisPage = lang.onThisPage;
      if (lang.totalPages) totalPages = lang.totalPages;
      document.getElementById('stat').innerHTML = `<br>${totalElemsFound}: ${filteredNum}<br>${onThisPage}: ${rowCount}<br>${totalPages}: ${totalPagesNum}`;  
    });
  } else {
    document.getElementById('stat').innerHTML = `<br>${totalElemsFound}: ${filteredNum}<br>${onThisPage}: ${rowCount}<br>${totalPages}: ${totalPagesNum}`;  
  }
}

function getElemsOnCurrentPage() {
  return document.getElementById('mainTable').getElementsByClassName('firstcell').length;
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
  const urlREGex = /([^a-z"'=]|^)((?:http|https|sftp|ftp):\/\/[^<\s\n\r\t]+[^.,;<\s\n\r\t"'])/gim;
  const wwwREGex = /([^a-z"'=/]|^)(www\.[^<\s\n\r\t]+[^.,;<\s\n\r\t"'])/gim;
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
        elem.innerHTML = elem.innerHTML.replace(urlREGex, '$1<a href="$2" target="_blank">$2</a>');
      }
      if (wwwREGex.test(elem.innerHTML)) {
        elem.innerHTML = elem.innerHTML.replace(wwwREGex, '$1<a href="http://$2" target="_blank">$2</a>');
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
