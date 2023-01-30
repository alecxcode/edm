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
      elem.innerHTML = elem.innerHTML
      .replace(/\[b\]/g, '<b>').replace(/\[\/b\]/g, '</b>')
      .replace(/\[i\]/g, '<i>').replace(/\[\/i\]/g, '</i>')
      .replace(/\[u\]/g, '<u>').replace(/\[\/u\]/g, '</u>')
      .replace(/\[q\]/g, '<q>').replace(/\[\/q\]/g, '</q>')
      .replace(/\[code\]/g, '<pre>').replace(/\[\/code\]/g, '</pre>');
      replaceNewlinesWithBR(elem);
      displayAsBlockQuoteIfBreaks(elem);
      if (urlREGex.test(elem.innerHTML)) {
        elem.innerHTML = elem.innerHTML.replace(urlREGex, '$1<a href="$2" target="_blank">$2</a>');
      }
      if (wwwREGex.test(elem.innerHTML)) {
        elem.innerHTML = elem.innerHTML.replace(wwwREGex, '$1<a href="http://$2" target="_blank">$2</a>');
      }
    }
  }
}

function replaceNewlinesWithBR(elem) {
  let childNodes = elem.childNodes;
  for (let node of childNodes) {
    if (node.nodeType == 3) {
      let textArr = node.nodeValue.replace(/\r/g, '').split('\n');
      const fragment = document.createDocumentFragment();
      for (let i = 0; i < textArr.length; i++) {
        if (i != 0) {
          const br = document.createElement('br');
          fragment.appendChild(br);
        }
        const text = document.createTextNode(textArr[i]);
        fragment.appendChild(text);
      }
      node.parentNode.replaceChild(fragment, node);
    } else if (node.nodeName != 'CODE' && node.nodeName != 'PRE') {
      replaceNewlinesWithBR(node);
    }
  }
}

function displayAsBlockQuoteIfBreaks(elem) {
  const quotes = elem.querySelectorAll('q');
  for (q of quotes) {
    if (q.innerHTML.includes('<br>') || q.innerHTML.length > 32) {
      q.style.display = 'block';
    }
  }
}

function clearBBCode(queryName) {
  let contentElems = document.querySelectorAll(queryName);
  contentElems.forEach(elem => {
    elem.innerHTML = elem.innerHTML
    .replace(/\[b\]/g, '').replace(/\[\/b\]/g, '')
    .replace(/\[i\]/g, '').replace(/\[\/i\]/g, '')
    .replace(/\[u\]/g, '').replace(/\[\/u\]/g, '')
    .replace(/\[q\]/g, '"').replace(/\[\/q\]/g, '"')
    .replace(/\[code\]/g, '').replace(/\[\/code\]/g, '');
  });
}