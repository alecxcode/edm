/* remove CSV Button if the page does not have grid data */
(function () {
  const csvButton = document.getElementById('makeCSV');
  const mainTable = document.getElementById('mainTable');
  const mainList = document.getElementById('mainList');
  if (csvButton && !(mainTable || mainList)) csvButton.parentNode.removeChild(csvButton);
})();

/* CSV make and download */
async function makeCSV(fileName) {
  let delim = ';';
  /* if (document.documentElement.lang != 'en') delim = ';'; */
  if (!fileName) {
    let pathArr = location.pathname.split('/');
    fileName = (pathArr.length > 0) ? pathArr[1] : 'export';
    let params = new URLSearchParams(window.location.search);
    let pageNumber = (params.has('pageNumber')) ? params.get('pageNumber') : '1';
    fileName = (fileName == 'export') ? fileName : fileName + "-" + pageNumber;
  }
  const mainTable = document.getElementById('mainTable');
  const mainList = document.getElementById('mainList');
  if (!mainTable && !mainList) {
    giveData('export', 'null');
    return;
  }
  if (mainTable) {
    let alldata = [];
    let headers = [];
    let row = [];
    let adding = false;
    document.querySelectorAll('.gridheader').forEach(h => headers.push(h.innerText));
    headers.shift();
    headers = headers.map(s => preProcessStrCSV(s.replace('▲', '').replace('▼', '')));
    alldata.push(headers.join(delim));
    let cells = mainTable.querySelectorAll("[class^=row]");
    if (cells) {
      cells.forEach(c => {
        if (c.classList.contains('firstcell')) adding = true;
        if (adding) row.push(preProcessStrCSV(c.innerText));
        if (c.classList.contains('lastcell')) {
          alldata.push(row.join(delim));
          row = [];
          adding = false;
        }
      });
    }
    giveData(fileName, alldata.join('\r\n'));
  }
  if (mainList && window.location.pathname.includes('compan')) {
    let alldata = [];
    let qarr = ['companyShort', 'companyFull', 'companyForeign', 'companyRegNo', 'companyTaxNo', 'companyAddrReg', 'companyHead'];
    let headers = ['', '', '', '', '', '', ''];
    let lang = await getLang(document.documentElement.lang);
    qarr.forEach((q, i) => { if (lang[q]) headers[i] = preProcessStrCSV(lang[q]); });
    console.log(headers);
    alldata.push(headers.join(delim));
    let rows = mainList.querySelectorAll('.panel');
    for (let r of rows) {
      let row = [];
      for (q of qarr) {
        let s = '';
        let e = r.querySelector(`[data-id="${q}"]`);
        if (e) s = preProcessStrCSV(e.innerText);
        row.push(s);
      }
      alldata.push(row.join(delim));
    }
    giveData(fileName, alldata.join('\r\n'));
  }
}

function preProcessStrCSV(s) {
  s = s.replace(/\r/g, '').replace(/\n/g, ' ').replace(/"/g, '""').trim();
  if (s.includes(',') || s.includes(';')) s = '"' + s + '"';
  return s;
}

/* Download a file */
function giveData(fileName, data) {
  let a = document.createElement('a');
  a.style.display = 'none';
  const blob = new Blob([data], { type: 'text/csv' });
  a.setAttribute('href', window.URL.createObjectURL(blob));
  a.setAttribute('download', fileName + '.csv');
  document.body.appendChild(a);
  a.click();
  document.body.removeChild(a);
}
