// https://vanillajstoolkit.com/helpers/sanitizehtml/
function sanitizeHTML (str) {
	return str.replace(/javascript:/gi, '').replace(/[^\w-_. ]/gi, function (c) {
		return `&#${c.charCodeAt(0)};`;
	});
}
