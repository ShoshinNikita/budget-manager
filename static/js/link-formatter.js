// We don't need perfect regexp for links. So, use this one without any strict checks for url correctness
const urlRegexp = /((https?:\/\/[^\s\/]+)[^\s]*)/g;

function formatLinks(querySelector) {
	const elements = document.querySelectorAll(querySelector);
	for (const elem of elements) {
		// Ignore all empty elements or elements with at least one child
		if (elem.children.length !== 0 || elem.innerHTML === "") continue;

		elem.innerHTML = elem.innerHTML.replace(urlRegexp, '<a href="$1" target="_blank" title="$1">$2/…</a>');
	}
}
