/* WebSocket related functions */
function connectWebSocket(reqObj, apiPath, callbackFn, onErrorCallbackFn) {
  let wsproto = 'ws';
  if (location.protocol == 'https:' || location.protocol == 'https') wsproto = 'wss';
  let ws = new WebSocket(`${wsproto}://${location.host}${apiPath}`);

  ws.onopen = () => {
    if (reqObj) ws.send(JSON.stringify(reqObj));
  }

  ws.onmessage = (event) => {
    if (event.data) callbackFn(JSON.parse(event.data));
  }

  var retriesCounter = 0;

  ws.onclose = () => {
    ws = null;
    var intervalID = setInterval(() => {
      retriesCounter++
      if (retriesCounter > 100000) clearInterval(intervalID);
      onErrorCallbackFn(intervalID);
    }, 8000);
  }

  ws.onerror = () => {
    ws = null;
    var intervalID = setInterval(() => {
      retriesCounter++
      if (retriesCounter > 100000) clearInterval(intervalID);
      onErrorCallbackFn(intervalID);
    }, 8000);
  }
}

function processOtherWSData(data) {
  if (data.error) console.log(data);
}
