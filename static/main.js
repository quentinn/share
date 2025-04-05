// Check if an URL exists
function checkUrl(url) {
	var http = new XMLHttpRequest();
	http.open('HEAD', url, false);
	http.send();
	return http.status == 200;
}



// Shortcut to get element
function element(element) {
	return document.getElementById(element);
}



// Display an error notification
function displayError(message) {
	let error = document.createElement('div');

	// Set the ID to always replace the same element with the error
	error.setAttribute("id", "error");

	error.innerHTML = message;
	error.className = 'notification';

	document.body.appendChild(error);

	// Automatically delete notification after few seconds
	setTimeout(function () {
		document.body.removeChild(error);
	}, 3000);

}



// Set logo from given file
function setLogo(fileLogo) {
	let logoImg = document.createElement('img');
	logoImg.setAttribute('src', fileLogo);

	let logo = document.getElementById('logo');
	logo.appendChild(logoImg)
}
