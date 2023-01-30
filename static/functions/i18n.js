/* Translates the page - aka i18n */
(function () {
  const langCode = document.documentElement.lang;
  if (langCode != 'en') {
    getLang(langCode).then(lang => {
      translateAll(null, lang);
      document.title = lang.appTitle + ': ' + document.querySelector('h1').innerText;
      const loadingDiv = document.querySelector('.fullscreen');
      if (loadingDiv) loadingDiv.parentNode.removeChild(loadingDiv);
      highlightCurrentBase();
    });
  }
})();

/* Translates recursively children elements of the root element */
function translateElem(rootElem) {
  const langCode = document.documentElement.lang;
  if (langCode != 'en') {
    getLang(langCode).then(lang => translateAll(rootElem, lang));
  }
}

/* Translates specific element only */
function translateCurrentElementOnly(elem) {
  if (document.documentElement.lang != 'en' && elem.getAttribute('i18n-text')) {
    getLang(document.documentElement.lang).then(lang => { if (lang[elem.getAttribute('i18n-text')]) elem.innerHTML = lang[elem.getAttribute('i18n-text')] });
  }
}

/* Puts text into element even if language is 'en' */
function putTextToCurrentElementOnly(elem) {
  if (elem.getAttribute('i18n-text')) {
    getLang(document.documentElement.lang).then(lang => { if (lang[elem.getAttribute('i18n-text')]) elem.innerHTML = lang[elem.getAttribute('i18n-text')] });
  }
}

/* Gets JSON language strings */
async function getLang(langCode) {
  const request = new Request(`/static/i18n/${langCode}.json`);
  const response = await fetch(request, { method: 'GET', credentials: 'include', mode: 'no-cors' });
  if (response.ok) {
    const lang = await response.json();
    return lang;
  } else {
    return {};
  }
}

/* Translates with provided JSON lang strings */
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
    if (c1 == ':' && c2 != '*' && s.charAt(s.length - 1) != ':') {
      s += ':';
    } else if (c1 == ':' && c2 == '*' && s.charAt(s.length - 2) != '*') {
      s += '*:';
    } else if (c1 == ' ' && c2 == ':') { s += ': '; }
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
