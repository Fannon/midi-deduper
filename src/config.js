export const defaultConfig = {

  //////////////////////////////////////////
  // MIDI Port Names                      //
  //////////////////////////////////////////

  instrumentInputPort: 'Finger Drum Pad',
  forwardPort1: 'loopMIDI Port',
  forwardPort2: '',

  //////////////////////////////////////////
  // General Options                      //
  //////////////////////////////////////////

  /**
   * How much time between notes need to pass at least, to not be a duplicate
   */
  timeThreshold: 60, 

  /**
   * How low the velocity needs to be to be detected as duplicate at all
   */
  velocityThreshold: 127,


  //////////////////////////////////////////
  // Advanced Options (no UI)             //
  //////////////////////////////////////////

  // How long the history is allowed to grow at most
  historyMaxSize: 25000, 
}

export function initConfig() {
  let config = defaultConfig
  const userConfig = localStorage.getItem("config");
  
  if (userConfig) {
    config = {
      ...config,
      ...JSON.parse(userConfig)
    }
  }

  updateSettingsInUI(config)

  console.debug('Config', config)

  return config
}

export function updateSettingsInUI(config) {

  document.getElementById('timeThreshold').value = config.timeThreshold.toString()
  document.getElementById('velocityThreshold').value = config.velocityThreshold.toString()

  // instrumentInputPort
  WebMidi.inputs.forEach((device) => {
    const option = document.createElement("option");
    option.text = device.name; 
    if (config.instrumentInputPort === device.name) {
      option.selected = true
    } 
    document.getElementById('instrumentInputPort').add(option)
  });

  // forwardPort1
  WebMidi.outputs.forEach((device) => {
    const option = document.createElement("option");
    option.text = device.name; 
    if (config.forwardPort1 === device.name) {
      option.selected = true
    } 
    document.getElementById('forwardPort1').add(option)
  });
  // forwardPort2
  WebMidi.outputs.forEach((device) => {
    const option = document.createElement("option");
    option.text = device.name; 
    if (config.forwardPort2 === device.name) {
      option.selected = true
    } 
    document.getElementById('forwardPort2').add(option)
  });
}

export function saveConfig(config, event) {
  if (event) {
    event.preventDefault() 
  }

  config.timeThreshold = parseInt(document.getElementById("timeThreshold").value);
  config.velocityThreshold = parseInt(document.getElementById("velocityThreshold").value);

  config.instrumentInputPort = document.getElementById("instrumentInputPort").value;
  config.forwardPort1 = document.getElementById("forwardPort1").value;
  config.forwardPort2 = document.getElementById("forwardPort2").value;

  localStorage.setItem("config", JSON.stringify(config));
  location.reload()
}

export function resetConfig(event) {
  if (event) {
    event.preventDefault() 
  }
  localStorage.removeItem("config")
  location.reload()
}
