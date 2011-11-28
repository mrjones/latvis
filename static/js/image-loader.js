var loadImage = function(filename, backoff) {
  if (backoff > 60) {
    document.getElementById('debug').innerHTML = 'Giving up.';    
  }
  doAjax("/is_ready/" + filename, function(result) {
    if (result == 'ok') {
      document.getElementById('spinner').style.display = 'none';
      var canvas = document.getElementById('canvas');
      var map = document.createElement('img');
      map.setAttribute('src', '/render/' + filename);
      canvas.appendChild(map);
    } else if (result = 'fail') {
      setTimeout("loadImage('" + filename + "', " + backoff + " * 1.5)", backoff * 1000);
      document.getElementById('debug').innerHTML = 'backing off: ' + backoff;
    } else {
      alert("Unexpected Result: " + result);
    }
  });
};


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
