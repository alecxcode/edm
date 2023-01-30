/* These functions are for elements creation and removal */
function makeElem(tag, parentElem, text, className, addSpace) {
  const elem = document.createElement(tag);
  if (className) elem.classList.add(className);
  if (text) elem.appendChild(document.createTextNode(text));
  parentElem.appendChild(elem);
  if (addSpace) parentElem.appendChild(document.createTextNode(' '));
  return elem;
}

function makeInputElem(type, parentElem, name, value, className, addSpace) {
  const elem = makeElem('input', parentElem, '', className, addSpace);
  if (type) elem.type = type;
  if (name) elem.name = name;
  if (value) elem.value = value;
  return elem;
}

/* Form creation and submitting */
function makeFormAndSubmit(id, elemName, action) {
  if (!id) return;
  let frm = document.createElement('form');
  frm.method = 'post';
  frm.action = action;
  let elem = document.createElement('input');
  elem.type = 'hidden';
  elem.name = elemName;
  elem.value = id;
  frm.appendChild(elem);
  document.body.appendChild(frm);
  frm.submit();
}

function clearChildNodes(parentElem) {
  while (parentElem.firstChild) {
    parentElem.removeChild(parentElem.firstChild);
  }
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
    const options = document.getElementById(inputID + 'List').children;
    for (let i = 0; i < options.length; i++) {
      if (options[i].getAttribute('data-value') == currencyCode) {
        inputCurrencyText.value = options[i].innerText;
        break;
      }
    }
  }
}

function addHiddenElem(theForm, name, value) {
  let input = document.createElement('input');
  input.type = 'hidden';
  input.name = name;
  input.value = value;
  theForm.appendChild(input);
}

function getTheme() {
  let links = document.getElementsByTagName('link');
  for (let i = 0; i < links.length; i++) {
    if (links[i].getAttribute('href').startsWith('/static/themes/')) {
      return links[i].getAttribute('href').replace('/static/themes/', '').replace('.css', '');
    }
    return '';
  }
}

function changeTheme(themeName) {
  let links = document.getElementsByTagName('link');
  for (let i = 0; i < links.length; i++) {
    if (links[i].getAttribute('href').startsWith('/static/themes/')) {
      let newHref = `/static/themes/${themeName}.css`;
      links[i].setAttribute('href', newHref);
      return;
    }
  }
}
