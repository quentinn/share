// Display an error notification
function display_error(message) {
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



function element(element) {
	return document.getElementById(element);
}