const localStorageKey = "theme";
const themeChangeEventName = "theme-change";
const darkThemeMedia = "(prefers-color-scheme: dark)";

const lightTheme = "light";
const systemTheme = "system";
const darkTheme = "dark";

/**
 * Saves a new theme in the local storage and updates the styles
 */
function switchTheme(newTheme) {
	if (newTheme !== lightTheme && newTheme !== systemTheme && newTheme !== darkTheme) {
		return;
	}
	localStorage.setItem(localStorageKey, newTheme);

	updateStyles();
}

function updateStyles() {
	let newTheme = lightTheme;
	if (isDarkTheme()) {
		newTheme = darkTheme;
	}

	document.getElementsByTagName("html")[0].setAttribute("data-theme", newTheme);

	window.dispatchEvent(new Event(themeChangeEventName));
}

/**
 * Returns true if:
 *   * the local theme is set to 'dark'
 *   * the local theme is set to 'system' and the system theme is set to 'dark'
 */
function isDarkTheme() {
	const localStorageTheme = getLocalStorageTheme();
	if (localStorageTheme !== systemTheme) {
		return localStorageTheme === darkTheme;
	}

	return getSystemTheme() === darkTheme;
}

function getLocalStorageTheme() {
	let value = localStorage.getItem(localStorageKey);
	if (!value) {
		// Use the system theme by default
		value = systemTheme;
		localStorage.setItem(localStorageKey, value);
	}
	return value;
}

function getSystemTheme() {
	if (window.matchMedia(darkThemeMedia).matches) {
		return darkTheme;
	}
	return lightTheme;
}

window.matchMedia(darkThemeMedia).addEventListener("change", function () {
	if (getLocalStorageTheme() !== systemTheme) {
		// Ignore changes
		return;
	}

	updateStyles();
});

// Update styles on the page load. Don't use 'window.addEventListener("load")' to avoid flickering
updateStyles();
