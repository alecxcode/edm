/* Replaces the page number arg in the URL according to last query after POST */
function replacePageNumberArg(pageNumber){
  let lastQuery = sessionStorage.getItem('lastQuery');
  if (lastQuery) {
    sessionStorage.removeItem('lastQuery');
    lastQuery = replaceGetArg(lastQuery, 'pageNumber', pageNumber);
    window.history.replaceState(null, '', lastQuery);
  }
}

/* Pagination mechanics */
function paginate(val) {
  const paginatorControl = document.getElementById('pageNumber');
  let page = Number(paginatorControl.value);
  let maxv = Number(paginatorControl.getAttribute("max"));
  if (val == 'max') val = maxv;
  if (val == 'min') { page = 1; val = 0 };
  page += val;
  if (page < 1) page = 1;
  if (page >= maxv) page = maxv;
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

function addStandardPagination(formID) {
  const frm = document.getElementById(formID);
  let pageNumber = document.getElementById('pageNumber').value;
  let elemsOnPage = document.getElementById('elemsOnPage').value;
  let elemsOnCurrentPage = getElemsOnCurrentPage();
  addHiddenElem(frm, 'elemsOnCurrentPage', elemsOnCurrentPage);
  addHiddenElem(frm, 'pageNumber', pageNumber);
  addHiddenElem(frm, 'elemsOnPage', elemsOnPage);
}

function addSeekPagination(formID, filteredNum) {
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
  addHiddenElem(frm, 'lastElemOnPage', mainTable.getElementsByClassName('firstcell')[mainTable.getElementsByClassName('firstcell').length - 1].lastElementChild.innerText);
}
