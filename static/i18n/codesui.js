(function(){
  
  /* define new added languages in this list: */
  let langCodes = {
    'en': 'English',
    'es': 'Español',
    'fr': 'Français',
    'ru': 'Русский',
  }

  const langSelector = document.getElementById('langCode');
  let selectOptions = langSelector.options;
  for (let opt of selectOptions) {
    if (langCodes[opt.value]) opt.innerHTML = langCodes[opt.value];
  }

})();