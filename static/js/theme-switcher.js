const localStorageKey = "theme";
const themeChangeEventName = "theme-change";

const lightTheme = "light";
const darkTheme = "dark";

function toggleTheme() {
	let newTheme = lightTheme;
	if (localStorage.getItem(localStorageKey) === lightTheme) {
		newTheme = darkTheme;
	}

	switchTheme(newTheme);
}

function switchTheme(newTheme) {
	const themeAttrName = "data-theme";

	localStorage.setItem(localStorageKey, newTheme);
	document.getElementsByTagName("html")[0].setAttribute(themeAttrName, newTheme);
	window.dispatchEvent(new Event(themeChangeEventName));
}

function isDarkTheme() {
	return localStorage.getItem(localStorageKey) === darkTheme;
}

// Call immediately
(function () {
	let theme = localStorage.getItem(localStorageKey);
	if (theme != lightTheme && theme != darkTheme) {
		// Use light theme as a default one
		theme = lightTheme;
	}
	switchTheme(theme);
})();
