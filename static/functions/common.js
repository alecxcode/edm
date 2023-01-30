/* Highlights current website block */
function highlightCurrentBase() {
  let addr = window.location.pathname.slice(0, 5);
  let currentItem;
  currentItem = document.querySelector(`#textmenu a[href^='${addr}']`);
  if (currentItem) currentItem.classList.add("chosenmenu");
};

(function () {
  if (document.documentElement.lang == 'en') highlightCurrentBase();
})();

/* This brings all sums to the format: [-]####.## on event */
function normalizeSum(elem) {
  setTimeout(function () {
    let s = elem.value.replace(/[^-0-9.,]+/g, "").replace(/[,.]+/g, ".");
    if (s) {
      elem.value = s[0] + s.slice(1).split("-").join("");
    } else {
      elem.value = s;
    }
  }, 200);
}

/* This brings all sums to the format: [-]# ###.## on load */
(function () {
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

/* This scrolls anchors to normal position so navbar does not overflow */
(function () {
  window.addEventListener('load', scrollToCorrectPosition, false);
  window.addEventListener('hashchange', scrollToCorrectPosition, false);
  function scrollToCorrectPosition() {
    const urlAnchorRegExp = /#([0-9A-Za-z]+)$/;
    let match = window.location.href.match(urlAnchorRegExp);
    if (match) {
      window.scrollTo(match[1].text);
      window.scrollBy(0, -44);
    }
  }
})();

/* Get the last part of the url pathname */
function getCurrentResourceID() {
  return +window.location.pathname.split('/').slice(-1)[0];
}

/* Replaces a value of an agrument in an URL get request */
function replaceGetArg(url, key, val) {
  let searchParams = new URLSearchParams(url.split('?')[1]);
  searchParams.set(key, val);
  return '?' + searchParams.toString();
}

/* Adds a value of an agrument to an URL get request */
function addGetArg(url, key, val) {
  let res = '';
  let rarr = url.split('?');
  if (rarr.length > 1) {
    res = url + '&' + key + '=' + val;
  } else {
    res = url + '?' + key + '=' + val;
  }
  return res;
}

/* Creates a user profile name from JSON subfield */
function makeProfileName(p) {
  if (!p) return '';
  let n = (p.FirstName + " " + p.Surname).trim();
  if (!n) {
    n = "ID: " + String(p.ID);
  }
  if (p.JobTitle) {
    n += ", " + p.JobTitle;
  }
  return n;
}

/* Check for newly created object */
function checkForNewCreated() {
  let result = sessionStorage.getItem('new');
  if (result) {
    sessionStorage.removeItem('new');
    if (!window.location.href.endsWith('new')) {
      const controlDiv = document.getElementById('control');
      const oldResDisplay = document.getElementById('resDisplay');
      if (oldResDisplay) controlDiv.removeChild(oldResDisplay);
      let msgnew = document.createElement('div');
      msgnew.className = "msgok";
      msgnew.innerHTML = "Creation completed.";
      msgnew.id = 'resDisplay';
      const langCode = document.documentElement.lang;
      if (langCode != 'en') {
        getLang(langCode).then(lang => { if (lang.creationCompleted) msgnew.innerHTML = lang.creationCompleted });
      }
      controlDiv.appendChild(msgnew);
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
    if (removal) removal.style.display = 'none';
    if (removalFiles) removalFiles.style.display = 'none'
  } else if (what == 'removal') {
    view.style.display = 'block';
    edit.style.display = 'none';
    if (removal) removal.style.display = 'block';
    if (removalFiles) removalFiles.style.display = 'none'
  } else if (what == 'removalFiles') {
    view.style.display = 'block';
    edit.style.display = 'none';
    if (removal) removal.style.display = 'none';
    if (removalFiles) removalFiles.style.display = 'block'
  } else {
    view.style.display = 'block';
    edit.style.display = 'none';
    if (removal) removal.style.display = 'none';
    if (removalFiles) removalFiles.style.display = 'none'
  }
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
    for (eachNode of document.getElementById('textmenu').childNodes) {
      d.appendChild(eachNode.cloneNode(true));
    }
  } else {
    document.getElementById('container').style.display = 'block';
    d.style.display = 'none';
    document.body.style.backgroundColor = styleRoot.getPropertyValue('--defaultBackgroundColor');
    clearChildNodes(d);
  }

}
