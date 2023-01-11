/* Common functions */

/* Function to make any request and call a callbackFn on data */
async function makeRequest(reqObj, apiURL, callbackFn) {
  const req = new Request(apiURL, {
    method: 'post',
    body: JSON.stringify(reqObj),
    headers: { 'Content-Type': 'application/json' }
  });
  fetch(req).then((response) => response.json())
  .catch((err) => {return {"error": 801, "description": err}})
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

function getCurrentResourceID() {
  return +window.location.pathname.split('/').slice(-1)[0];
}

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



/* Project management functions */
function setProjStatus(status, projStatusesNamesList) {
  makeRequest({proj: getCurrentResourceID(), status: status},
    '/projs/setstatus',
    (project) => {
    resDisplayUpdate('dataWritten');
    const statusIndicator = document.getElementById('statusIndicator');
    statusIndicator.innerText = giveStatus(project, projStatusesNamesList);
    statusIndicator.setAttribute('i18n-index', 'projStatuses-'+project.ProjStatus);
    if (document.documentElement.lang != 'en') {
      getLang(document.documentElement.lang).then(lang => {if (lang.projStatuses) statusIndicator.innerHTML = lang.projStatuses[project.ProjStatus]});
    }
    updateProjStatusIndicatorClass(project.ProjStatus);
    updateProjStatusButtons(project.ProjStatus);
  });
}

function giveStatus(project, projStatusesNamesList) {
	if (project.ProjStatus < projStatusesNamesList.length && project.ProjStatus >= 0) {
		return projStatusesNamesList[project.ProjStatus]
	} else {
		return "Unknown"
	}
}

function updateProjStatusIndicatorClass(status) {
  const statusIndicator = document.getElementById('statusIndicator');
  if (!statusIndicator) return;
  const ACTIVE = 0, DONE = 1, CANCELED = 2;
  let className = '';
  if (status == ACTIVE) className = 'txtblue';
  if (status == DONE) className = 'txtgreen';
  if (status == CANCELED) className = 'txtbw';
  if (className) statusIndicator.className = className;
}

function updateProjStatusButtons(status) {
  const statusArrTxtIDs = ['active', 'done', 'canceled'];
  statusArrTxtIDs.forEach(textID => {
    let button = document.getElementById(textID);
    if (button) button.disabled = false;
  });
  let textStatus = statusArrTxtIDs[status];
  let textStatusButton;
  if (textStatus) textStatusButton = document.getElementById(textStatus);
  if (textStatusButton) textStatusButton.disabled = true;
}



/* Search on project page*/
function searchOnPage(textFilter) {
  textFilter = textFilter.toLowerCase().replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;');
  let searchRoots = document.getElementsByClassName('elem-inside-item');
  if (!textFilter) searchOnPageReset();
  if (textFilter && searchRoots) {
    for (content of searchRoots) {
      let elems = content.getElementsByTagName('a');
      if (elems) {
        for (let elem of elems) {
          elem.innerHTML = elem.innerHTML.replace(/<span class="highlight">/g, '').replace(/<\/span>/g, '');
          recursiveChildrenSearch(elem, textFilter);
        }
      }
      if (content.querySelector('span.highlight')) {
        content.style.display = 'block';
      } else {
        content.style.display = 'none';
      }
    }
  }
}

function searchOnPageReset() {
  let searchRoots = document.getElementsByClassName('elem-inside-item');
  if (searchRoots) {
    for (content of searchRoots) {
      let elems = content.getElementsByTagName('a');
      if (elems) {
        for (let elem of elems) {
          elem.innerHTML = elem.innerHTML.replace(/<span class="highlight">/g, '').replace(/<\/span>/g, '');
        }
      }
      content.style.display = 'block';
    }
  }
}



/* Task management functions */

/* Tasks loading */
function loadTasks(userRights) {
  if (!window.location.pathname.endsWith('/new')) {
    makeRequest({name: "project", id: getCurrentResourceID()},
      '/tasks/loadtasks',
      (taskList) => {taskList.forEach((task) => {
        putTaskIntoCol(task, false, userRights);
      });
      reCalcTaskNum();
      buildProjectParticipants();
    });
  }
}

/* Task project adding */
function addTaskToProject(taskID, userRights) {
  if (!document.getElementById(taskID)) {
    makeRequest({task: taskID, proj: getCurrentResourceID()}, 
      '/tasks/updateproj',
      (task) => {
        resDisplayUpdate('dataWritten');
        putTaskIntoCol(task, true, userRights);
        reCalcTaskNum();
        buildProjectParticipants();
      });
  }
}

/* Task project removing */
function detachTaskFromProject(taskID) {
  if (document.getElementById(taskID)) {
    makeRequest({task: taskID, proj: null},
      '/tasks/updateproj',
      (task) => {
        resDisplayUpdate('dataWritten');
        if (document.getElementById(task.ID)) {
          document.getElementById(task.ID).remove();
        }
        reCalcTaskNum();
        buildProjectParticipants();
      });
  }
}

/* Task assigning */
function assignTaskToDo(taskID, assigneeID) {
  makeRequest({task: taskID, assignee: assigneeID}, 
    '/tasks/assigntask',
    (task) => {
      if (document.getElementById(task.ID)) {
        document.getElementById(task.ID).remove();
      }
      resDisplayUpdate('dataWritten');
      putTaskIntoCol(task, true);
      reCalcTaskNum();
      buildProjectParticipants();
    });
}

/* Show task project on task page */
function showProjectOnTaskPage(project) {
  const pa = document.querySelector(`a[href^='/projs/project/']`);
  if (pa) pa.innerText = project.ProjName;
  const projectBlock = document.getElementById('projectBlock');
  if (projectBlock) {
    const spanOfName = makeElem('span', projectBlock, 'Project:', 'ofname', true);
    spanOfName.setAttribute('i18n-text', 'project');
    const inpSpan = makeElem('span', projectBlock, '', 'nowrap', true);
    const cb = makeInputElem('checkbox', inpSpan, 'project', project.ID, '', true);
    cb.checked = true;
    cb.id = 'pidcb';
    const plabel = makeElem('label', inpSpan, project.ProjName, '', false);
    plabel.setAttribute('for', cb.id);
    translateElem(projectBlock);
    projectBlock.style.display = 'block';
  }
}

/* Select task column */
function selectTaskColumn(status) {
  const CREATED = 0, ASSIGNED = 1, INPROGRESS = 2, STUCK = 3, DONE = 4, CANCELED = 5, INREVIEW = 6;
  let col;
  if (status == CREATED) {
    col = document.getElementById('col-created');
  } else if (status == ASSIGNED || status == INPROGRESS) {
    col = document.getElementById('col-inprogress');
  } else if (status == STUCK || status == INREVIEW) {
    col = document.getElementById('col-attn');
  } else if (status == DONE || status == CANCELED) {
    col = document.getElementById('col-completed');
  }
  if (!col) col = document.getElementById('col-attn');
  return col;
}

function calcPosition(id, arr) {
  let res = arr[0];
  for (let i = 0; i < arr.length; i++) {
    res = arr[i];
    if (id < arr[i]) {
      return res;
    }
  }
  return false;
}

function updateTaskStatusIndicatorClass(status, elem) {
  let statusIndicator;
  if (elem){
    statusIndicator = elem;
  } else {
    statusIndicator = document.getElementById('statusIndicator');
  }
  if (!statusIndicator) return;
  const CREATED = 0, ASSIGNED = 1, INPROGRESS = 2, STUCK = 3, DONE = 4, CANCELED = 5, INREVIEW = 6;
  let className = '';
  if (status == CREATED || status == ASSIGNED || status == INPROGRESS) className = 'txtblue';
  if (status == STUCK) className = 'txtred';
  if (status == DONE) className = 'txtgreen';
  if (status == CANCELED) className = 'txtbw';
  if (status == INREVIEW) className = 'txtora';
  if (className) {
    if (statusIndicator.className.includes('txt')) {
      statusIndicator.className = className;
    } else {
      statusIndicator.classList.add(className);
    }
  }
}

function updateTaskStatusButtons(status) {
  const statusArrTxtIDs = ['', '', 'inprogress', 'stuck', 'done', 'canceled', 'inreview'];
  statusArrTxtIDs.forEach((textID, i) => {
    if (i >= 2) {
      let button = document.getElementById(textID);
      if (button) button.disabled = false;
    }
  });
  let textStatus = statusArrTxtIDs[status];
  let textStatusButton;
  if (textStatus) textStatusButton = document.getElementById(textStatus);
  if (textStatusButton) textStatusButton.disabled = true;
}

/* Create and insert task element */
function putTaskIntoCol(task, placeUp, userRights) {
  const CREATED = 0;
  const col = selectTaskColumn(task.TaskStatus);
  if (document.getElementById(task.ID)) {
    if (document.getElementById(task.ID).parentNode == col) return;
  }

  const elem = makeElem('div', col, '', 'elem-inside-item', false);
  elem.id = task.ID;

  if (placeUp) {
    divs = col.querySelectorAll('div.elem-inside-item');
    let arrIDs = [];
    for (eachDiv of divs) {
      arrIDs.push(Number(eachDiv.id));
    }
    let idBefore = calcPosition(task.ID, arrIDs);
    if (idBefore) {
      let targetElem = document.getElementById(idBefore);
      if (targetElem) col.insertBefore(elem, targetElem);
    }
  }
  
  const topic = makeElem('a', elem, `#${task.ID}: ` + task.Topic, 'bigger-font', false);
  topic.setAttribute('href', `/tasks/task/${task.ID}`);

  const statusesArr = ["Created", "Assigned", "In progress", "Stuck", "Done", "Canceled", "In review"];
  const statusDiv = makeElem('div', elem, '', '', false);
  statusDiv.setAttribute('data-status', task.TaskStatus);
  const statusLabel = makeElem('span', statusDiv, 'Status:', '', true);
  statusLabel.setAttribute('i18n-text', 'status');
  const statusIndicator = makeElem('span', statusDiv, statusesArr[task.TaskStatus], '', false);
  statusIndicator.setAttribute('i18n-index', `taskStatuses-${task.TaskStatus}`);
  updateTaskStatusIndicatorClass(task.TaskStatus, statusIndicator);

  if (task.Creator) {
    const creatorDiv = makeElem('div', elem, '', '', false);
    const creatorLabel = makeElem('span', creatorDiv, 'Creator (owner):', '', true);
    creatorLabel.setAttribute('i18n-text', 'creatorOwner');
    const creatorName = makeElem('a', creatorDiv, makeProfileName(task.Creator), '', false);
    creatorName.setAttribute('href', `/team/profile/${task.Creator.ID}`);
  }

  if (task.Assignee) {
    const assigneeDiv = makeElem('div', elem, '', '', false);
    const assigneeLabel = makeElem('span', assigneeDiv, 'Assignee:', '', true);
    assigneeLabel.setAttribute('i18n-text', 'assignee');
    const assigneeName = makeElem('a', assigneeDiv, makeProfileName(task.Assignee), '', false);
    assigneeName.setAttribute('href', `/team/profile/${task.Assignee.ID}`);
    assigneeName.setAttribute('data-id', task.Assignee.ID);
  }

  if (task.TaskStatus == CREATED) {
    const inpTake = makeInputElem('button', elem, 'takeButton', 'Take', 'sbut', true);
    inpTake.classList.add('greenish', 'smaller');
    inpTake.setAttribute('i18n-many', 'take');
    inpTake.addEventListener('click', (e) => {assignTaskToDo(+e.target.parentNode.id, null)});  
  }

  let createAssignForm = false;
  if (userRights) {
    if (userRights.loggedinAdmin) {
      createAssignForm = true;
    } else if (task.Creator && userRights.loggedinID == task.Creator.ID) {
      createAssignForm = true;
    } else if (task.Assignee && userRights.loggedinID == task.Assignee.ID) {
      createAssignForm = true;
    } else if (userRights.loggedinID == userRights.projectCreator) {
      createAssignForm = true;
    }
  }

  if (task.TaskStatus == CREATED && createAssignForm) {
    const inpAssign = makeInputElem('button', elem, 'assignButton', 'Forward', 'sbut', true);
    inpAssign.classList.add('smaller');
    inpAssign.setAttribute('i18n-value', 'forward');
    inpAssign.addEventListener('click', (e) => {
      const selectedTaskInput = document.getElementById('selectedTask');
      const forwardWindow = document.getElementById('forwardWindow');
      const taskNoAndTopic = document.getElementById('taskNoAndTopic');
      if (forwardWindow) forwardWindow.style.display = 'block';
      if (selectedTaskInput) selectedTaskInput.value = e.target.parentNode.id;
      if (taskNoAndTopic) taskNoAndTopic.innerText = `#${task.ID}: ` + task.Topic;
    });
  }

  translateElem(elem);
}

function reCalcTaskNum(){
  const arr = ['created', 'inprogress', 'attn', 'completed'];
  arr.forEach(colName => {
    document.getElementById('num-'+colName).innerText = " (" + String(document.getElementById('col-'+colName).childElementCount - 1) + ")";
  });
}

function buildProjectParticipants() {
  const participantsDisplay = document.getElementById('participantsList');
  participantsDisplay.querySelectorAll('*').forEach(e => e.remove());
  participantsDisplay.innerHTML = '';
  let counter = 0;
  let uniqueMap = new Map();
  let sourceArr = document.querySelectorAll(`.item-col-4 a[href^='/team/profile/']`);
  sourceArr.forEach((a) => {
    uniqueMap.set(a.innerText, a.href.split('/').slice(-1)[0]);
  });
  let keys = [...uniqueMap.keys()].sort();
  keys.forEach((profileName) => {
    const creatorName = makeElem('a', participantsDisplay, profileName, 'chosen', true);
    creatorName.setAttribute('href', `/team/profile/${uniqueMap.get(profileName)}`);
    counter++;
  });
  
  const participantsNum = document.getElementById('participantsNum');
  participantsNum.innerText = '(' + counter +')';

  if (counter == 0) {
    const noParticipants = makeElem('span', participantsDisplay, 'no', '', false);
    noParticipants.setAttribute('i18n-text', 'no');
    translateCurrentElementOnly(noParticipants);
  }
}



/* Other functions */
function resDisplayUpdate(message) {
  const oldResDisplay = document.getElementById('resDisplay');
  const controlDiv = document.getElementById('control');
  controlDiv.removeChild(oldResDisplay);
  const resDisplay = makeElem('div', controlDiv, '', '', false);
  resDisplay.id = 'resDisplay';

  if (message.error) {
    resDisplay.className = 'msgred';
  }
  if (message.error && message.error == 804) {
    resDisplay.setAttribute('i18n-text', 'objectNotFound');
    resDisplay.innerHTML = 'Object not found.';
  } else if (message.error && message.error == 403) {
    resDisplay.setAttribute('i18n-text', 'noPerms');
    resDisplay.innerHTML = 'Action rejected: no permissions for all or some of objects.';
  } else if (message.error && message.error == 500) {
    resDisplay.setAttribute('i18n-text', 'serverError');
    resDisplay.innerHTML = 'Internal server error.';
  } else if (message.error) {
    resDisplay.setAttribute('i18n-text', 'unknownError');
    resDisplay.innerHTML = 'Unknown error.';
  } else if (message == "dataWritten") {
    resDisplay.setAttribute('i18n-text', message);
    resDisplay.className = 'msgok';
    resDisplay.innerHTML = 'Data has been written successfully.';
  } else if (message == "dataNotWritten") {
    resDisplay.setAttribute('i18n-text', message);
    resDisplay.className = 'msgred';
    resDisplay.innerHTML = 'Error writing data.';
  } else {
    resDisplay.className = ''
    resDisplay.innerHTML = '<br>';
  }

  translateCurrentElementOnly(resDisplay);
}

function translateCurrentElementOnly(elem) {
  if (document.documentElement.lang != 'en' && elem.getAttribute('i18n-text')) {
    getLang(document.documentElement.lang).then(lang => {if (lang[elem.getAttribute('i18n-text')]) elem.innerHTML = lang[elem.getAttribute('i18n-text')]});
  }
}



/* WebSocket related functions */
function connectWebSocket() {
  let ws = new WebSocket(`ws://${location.host}/projs/ws`);

  ws.onopen = () => {
    ws.send(JSON.stringify({"project": getCurrentResourceID()}));
  }

  ws.onmessage = (event) => {
    if (event.data) processArrivedTask(JSON.parse(event.data));
  }

  var retriesCounter = 0;

  ws.onclose = () => {
    ws = null;
    var intervalID = setInterval(() => {
      retriesCounter++
      if (retriesCounter > 100000) clearInterval(intervalID);
      tryToConnect(intervalID);
    }, 8000);
  }

  ws.onerror = () => {
    ws = null;
    var intervalID = setInterval(() => {
      retriesCounter++
      if (retriesCounter > 100000) clearInterval(intervalID);
      tryToConnect(intervalID);
    }, 8000);
  }
}

async function tryToConnect(intervalID) {
  req = new Request('/tasks/loadtasks', {
    method: 'post',
    body: JSON.stringify({name: "project", id: getCurrentResourceID()}),
    headers: { 'Content-Type': 'application/json' }
  });
  await fetch(req).then((response) => response.json())
  .catch((err) => {return {"error": 801, "description": err}})
  .then((data) => {
    if (!data.error) {
      console.log(data);
      let userRights = JSON.parse(sessionStorage.getItem("userRights"));
      data.forEach((task) => {
        putTaskIntoCol(task, false, userRights);
      });
      reCalcTaskNum();
      buildProjectParticipants();
      clearInterval(intervalID);
      connectWebSocket();
    } else {
      console.log(data);
    }
  });
}

function processArrivedTask(task) {

  if (!task.ID) {
    processOtherWSData(task);
    return
  }

  // Removing task
  if (task.Project == 0 || task.Project != getCurrentResourceID() || !task.Project) {
    if (document.getElementById(task.ID)) {
      document.getElementById(task.ID).remove();
      reCalcTaskNum();
      buildProjectParticipants();
    }
    return;
  }

  // Adding task
  if (task.Project == getCurrentResourceID() && !document.getElementById(task.ID)) {
    putTaskIntoCol(task, true, JSON.parse(sessionStorage.getItem("userRights")));
    reCalcTaskNum();
    buildProjectParticipants();
    return;
  }

  // Updating task
  if (task.Project == getCurrentResourceID() && document.getElementById(task.ID)) {
    const taskDiv = document.getElementById(task.ID);
    const assigneeLink = taskDiv.querySelector('a[data-id]');
    const statusDiv = taskDiv.querySelector('div[data-status]');
    let assigneeID = -1;
    let taskAssigneeID = -1;
    if (assigneeLink) assigneeName = Number(assigneeLink.getAttribute('data-id'));
    if (task.Assignee) taskAssigneeID = task.Assignee.ID;

    let currentStatus = -1;
    if (statusDiv) currentStatus = Number(statusDiv.getAttribute('data-status'));

    if (taskDiv.parentNode == selectTaskColumn(task.TaskStatus) &&
      assigneeID == taskAssigneeID && currentStatus == task.TaskStatus) {
      return;
    }

    document.getElementById(task.ID).remove();
    putTaskIntoCol(task, true, JSON.parse(sessionStorage.getItem("userRights")));
    reCalcTaskNum();
    buildProjectParticipants();
    return;
  }
}

function processOtherWSData(data) {
  if (data.error) console.log(data);
}
