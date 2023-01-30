/* Function to make any JSON request and call a callbackFn on data */
async function makeRequest(reqObj, apiURL, callbackFn) {
  const req = new Request(apiURL, {
    method: 'post',
    body: JSON.stringify(reqObj),
    headers: { 'Content-Type': 'application/json' }
  });
  fetch(req).then((response) => response.json())
  .catch((err) => { return { "error": 801, "description": err } })
  .then((data) => {
    if (!data) {
      return;
    } else if (data.error) {
      console.log(data);
      resDisplayUpdate(data);
    } else if (data && callbackFn) {
      callbackFn(data);
    }
  });
}

/* Function to make any non-json request and call a callbackFn on data */
async function makeDataRequest(reqObj, apiURL, callbackFn, targProp) {
  reqObj.append("api", "json");
  const req = new Request(apiURL, {
    method: 'post',
    body: reqObj
  });
  fetch(req).then((response) => response.json())
  .catch((err) => { return { "error": 801, "description": err } })
  .then((data) => {
    if (!data) {
      return;
    } else if (data.error) {
      console.log(data);
      resDisplayUpdate(data);
    } else if (data && callbackFn) {
      if (targProp) {
        callbackFn(data[targProp]);
        if (data.Message) resDisplayUpdate(data.Message);
      } else {
        callbackFn(data);
        if (data.Message) resDisplayUpdate(data.Message);
      }
    }
  });
}

/* Other functions */
function resDisplayUpdate(message) {
  const oldResDisplay = document.getElementById('resDisplay');
  const controlDiv = document.getElementById('control');
  controlDiv.removeChild(oldResDisplay);
  const resDisplay = makeElem('div', controlDiv, '', '', false);
  resDisplay.id = 'resDisplay';

  function isMessageErrory() {
    if (message.error) return true;
    if (typeof message != 'string') return false;
    let arrOfErr = ['Error', 'NotWritten', 'AlreadyInList', 'noData', 'NotFound', 'noPerms', 'removalPerms'];
    for (let s of arrOfErr) {
      if (message.includes(s)) return true;
    }
    return false;
  }

  if (isMessageErrory(message)) {
    resDisplay.className = 'msgred';
  } else {
    resDisplay.className = 'msgok';
  }

  if (message.error && message.error == 804) {
    resDisplay.setAttribute('i18n-text', 'objectNotFound');
  } else if (message.error && message.error == 403) {
    resDisplay.setAttribute('i18n-text', 'noPerms');
  } else if (message.error && message.error == 500) {
    resDisplay.setAttribute('i18n-text', 'serverError');
  } else if (message.error) {
    resDisplay.setAttribute('i18n-text', 'unknownError');
  } else if (typeof message == 'string') {
    resDisplay.setAttribute('i18n-text', message);
  } else {
    resDisplay.removeAttribute('i18n-text');
    resDisplay.className = ''
    resDisplay.innerHTML = '<br>';
  }

  putTextToCurrentElementOnly(resDisplay);
}
