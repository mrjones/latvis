var loadImage = function(filename, backoff) {
  if (backoff > 60) {
    document.getElementById('debug').innerHTML = 'Giving up.';    
  }
  doAjax("/is_ready/" + filename, function(result) {
    if (result == 'ok') {
      document.getElementById('loading').style.display = 'none';
      var canvas = document.getElementById('canvas');
      var map = document.createElement('img');
      map.setAttribute('src', '/render/' + filename);
      canvas.appendChild(map);
      renderMetadata();
      _gat._getTrackerByName()._trackEvent("latvis-render", "render-complete");
    } else if (result = 'fail') {
      setTimeout("loadImage('" + filename + "', " + backoff + " * 1.5)", backoff * 1000);
      _gat._getTrackerByName()._trackEvent("latvis-render", "timeout-error");
//      document.getElementById('debug').innerHTML = 'backing off: ' + backoff;
    } else {
      alert("Unexpected Result: " + result);
    }
  });
};

function renderMetadata() {
  var container = document.getElementById('metadata');
  var url = window.location;

  container.innerHTML = "<span class='latvis-generate-link'><a href='http://latvis.mrjon.es'>Generate new image</a></span> | <a href='" + url + "'>Link to this page</a> | <a href='https://twitter.com/share'>Tweet</a>";

  container.style.display = 'block';
}

function doAjax(url, handler) {
//  document.getElementById('debug').innerHTML = 'Making AJAX call...';
  var xmlHttpReq = false;
  var self = this;
  // Mozilla/Safari
  if (window.XMLHttpRequest) {
    self.xmlHttpReq = new XMLHttpRequest();
  }
  // IE
  else if (window.ActiveXObject) {
    self.xmlHttpReq = new ActiveXObject("Microsoft.XMLHTTP");
  }
  self.xmlHttpReq.open('POST', url, true);
  self.xmlHttpReq.setRequestHeader('Content-Type', 'application/x-www-form-urlencoded');
  self.xmlHttpReq.onreadystatechange = function() {
    if (self.xmlHttpReq.readyState == 4) {
      handler(self.xmlHttpReq.responseText);
    }
  }
  self.xmlHttpReq.send();
};
