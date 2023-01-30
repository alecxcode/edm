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
  visibleElem.className = 'chosen-block';
  visibleElem.appendChild(document.createTextNode(profileName));
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
    filterboxes.forEach(function (cb) {
      cb.checked = (arrayOfTypes.includes(+cb.value)) ? true : false;
    });
  }
}

/* This function handles date filters applied from server */
function applyDateFilter(dateFilter) {
  //dateFilter.DatesStr.map(v => v.replace(' ', 'T'));
  if (dateFilter.DatesStr.length == 1) {
    document.getElementById(dateFilter.Name + 'Single').value = dateFilter.DatesStr[0];
    document.getElementById(dateFilter.Name + 'Relation').value = dateFilter.Relation;
  }
  if (dateFilter.DatesStr.length == 2) {
    const cb = document.getElementById(dateFilter.Name + 'Diapason');
    cb.checked = true;
    handleNumericFilterChkBox(cb, dateFilter.Name);
    handleNumericOption(cb, dateFilter.Name);
    document.getElementById(dateFilter.Name + 'Start').value = dateFilter.DatesStr[0];
    document.getElementById(dateFilter.Name + 'Finish').value = dateFilter.DatesStr[1];
  }
}

/* This function handles sum filters applied from server */
function applySumFilter(sumFilter) {
  const filterCurrency = document.getElementById(sumFilter.Name + 'Currency');
  const filterCurrencyCode = document.getElementById(sumFilter.Name + 'CurrencyCode');
  if (filterCurrency) {
    const options = document.getElementById(sumFilter.Name + 'CurrencyCodesList').children;
    for (let i = 0; i < options.length; i++) {
      if (options[i].getAttribute('data-value') == sumFilter.CurrencyCode) {
        filterCurrency.value = options[i].innerText;
        filterCurrencyCode.value = options[i].getAttribute('data-value');
        break;
      }
    }
  }
  if (sumFilter.Sums.length == 1) {
    document.getElementById(sumFilter.Name + 'Single').value = sumFilter.SumsStr[0];
    document.getElementById(sumFilter.Name + 'Relation').value = sumFilter.Relation;
  }
  if (sumFilter.Sums.length == 2) {
    const cb = document.getElementById(sumFilter.Name + 'Diapason');
    cb.checked = true;
    handleNumericFilterChkBox(cb, sumFilter.Name);
    handleNumericOption(cb, sumFilter.Name);
    document.getElementById(sumFilter.Name + 'Start').value = sumFilter.SumsStr[0];
    document.getElementById(sumFilter.Name + 'Finish').value = sumFilter.SumsStr[1];
  }
}

function handleNumericFilterChkBox(cb, elemIDPrefix) {
  if (cb.checked == true) {
    document.getElementById(elemIDPrefix + 'Single').disabled = true;
    document.getElementById(elemIDPrefix + 'Start').disabled = false;
    document.getElementById(elemIDPrefix + 'Finish').disabled = false;
  } else {
    document.getElementById(elemIDPrefix + 'Single').disabled = false;
    document.getElementById(elemIDPrefix + 'Start').disabled = true;
    document.getElementById(elemIDPrefix + 'Finish').disabled = true;
  }
}
function handleNumericOption(cb, elemIDPrefix) {
  if (cb.checked == true) {
    document.getElementById(elemIDPrefix + 'Relation').disabled = true;
  } else {
    document.getElementById(elemIDPrefix + 'Relation').disabled = false;
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
  textFilter = textFilter.toLowerCase().replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
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
    elem.innerHTML = insertCaseInsensitive(elem.innerHTML, textFilter, '<span class="highlight">', '</span>');
  }
}

/* Insert a highlight tag */
function insertCaseInsensitive(srcStr, lowerCaseFilter, before, after) {
  let lowStr = srcStr.toLowerCase();
  let flen = lowerCaseFilter.length;
  let i = -1;
  while ((i = lowStr.indexOf(lowerCaseFilter, i + 1)) != -1) {
    srcStr = srcStr.slice(0, i) + before + srcStr.slice(i, i + flen) + after + srcStr.slice(i + flen);
    lowStr = srcStr.toLowerCase();
    i += before.length + after.length;
  }
  return srcStr;
}



/* This function prints message about applied filters */
function printAppliedFilters(classFilterArr, dateFilterArr, sumFilterArr, textFilter) {

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

  var intervalID = setInterval(function () {
    retries++
    if (retries > 10000) clearInterval(intervalID);
    if (!document.querySelector('.fullscreen') && runAllowed) {
      clearInterval(intervalID);
      const fragment = document.createDocumentFragment();
      for (let classFilter of classFilterArr) {
        if (classFilter.Selector) {
          const printedName = document.getElementById(classFilter.Name + 'Display').innerText;
          makeElem('span', fragment, printedName, 'af', true);
          let namesArr = document.getElementById(classFilter.Name).children;
          for (let i = 1; i < namesArr.length; i++) {
            if (namesArr[i].innerText && namesArr[i].innerText != "\n" && namesArr[i].innerText != " ") {
              makeElem('span', fragment, namesArr[i].innerText, 'af', true);
            }
          }
          makeElem('br', fragment);
        } else {
          const filterboxes = document.getElementsByName(classFilter.Name);
          filterboxes.forEach(function (cb) {
            if (cb.checked) {
              const filterLabel = document.querySelector(`label[for="${cb.id}"]`);
              makeElem('span', fragment, filterLabel.innerText, 'af', true);
            }
          });
        }
      }
      for (let dateFilter of dateFilterArr) {
        dateFilter.DatesStr = dateFilter.DatesStr.map(v => v.replace('T', ' '));
        const printedName = document.getElementById(dateFilter.Name + 'Display').innerText;
        let rel = getSelectInputText(document.getElementById(dateFilter.Name + 'Relation'));
        if (dateFilter.DatesStr.length == 1) {
          makeElem('span', fragment, `${printedName} ${rel}: ${dateFilter.DatesStr[0]}`, 'af', true);
        }
        if (dateFilter.DatesStr.length == 2) {
          makeElem('span', fragment, `${printedName} ${diapason}: ${dateFilter.DatesStr[0]} - ${dateFilter.DatesStr[1]}`, 'af', true);
        }
      }
      for (let sumFilter of sumFilterArr) {
        let currency = document.getElementById(sumFilter.Name + 'Currency').value;
        if (sumFilter.CurrencyCode > 0) makeElem('span', fragment, currency, 'af', true);
        const printedName = document.getElementById(sumFilter.Name + 'Display').innerText;
        let rel = getSelectInputText(document.getElementById(sumFilter.Name + 'Relation'));
        if (sumFilter.Sums.length == 1) {
          makeElem('span', fragment, `${printedName} ${rel}: ${sumFilter.SumsStr[0]}`, 'af', true);
        }
        if (sumFilter.Sums.length == 2) {
          makeElem('span', fragment, `${printedName} ${diapason}: ${sumFilter.SumsStr[0]} - ${sumFilter.SumsStr[1]}`, 'af', true);
        }
      }
      if (textFilter) {
        makeElem('span', fragment, textFilter, 'af', true);
      }
      const p = document.getElementById('appliedFilters');
      clearChildNodes(p);
      if (fragment.children.length > 0) {
        p.appendChild(document.createTextNode(appliedFilters));
        p.appendChild(fragment);
      } else {
        p.appendChild(document.createTextNode(noFiltersApplied));
      }
    }
  }, timeout);

}



/* Resets filters */
function resetFilter() {
  const frm = document.getElementById('controlForm');
  frm.reset();
  removeEmptyInputs(frm);
  addSortingOnSubmut('controlForm');
  frm.submit();
}

/* Add filters inputs to a form */
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
      addHiddenElem(frm, eachF.Name + 'Relation', eachF.Relation);
      for (let i = 0; i < eachF.Dates.length; i++) {
        addHiddenElem(frm, eachF.Name, eachF.DatesStr[i]);
      }
    }
  }

  if (sumFilter) {
    for (let eachF of sumFilter) {
      addHiddenElem(frm, eachF.Name + 'CurrencyCode', eachF.CurrencyCode);
      addHiddenElem(frm, eachF.Name + 'Relation', eachF.Relation);
      for (let i = 0; i < eachF.SumsStr.length; i++) {
        addHiddenElem(frm, eachF.Name, eachF.SumsStr[i]);
      }
    }
  }

  if (textFilter) {
    addHiddenElem(frm, 'searchText', textFilter);
  }
}
