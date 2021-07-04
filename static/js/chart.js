// Update global Chart.js options
(function () {
	const global = Chart.defaults;
	const plugins = global.plugins;

	// Hide legend
	plugins.legend.display = false;
	// Disable animations
	global.animation.duration = 0;
	// Tune tooltips
	plugins.tooltip.animation.duration = 200;
	plugins.tooltip.titleFontSize = 15;
	plugins.tooltip.backgroundColor = "#000000d0";
	plugins.tooltip.cornerRadius = 5;
	// Other
	global.font.size = 14;
	global.maintainAspectRatio = false;
	// Tune scale
	const scale = Chart.defaults.scale;
	scale.ticks.beginAtZero = true;
	scale.grid.color = getGridLinesColor();

	// Register custom formatter for money
	Chart.Ticks.formatters.money = formatMoney;
})();

function getGridLinesColor() {
	return isDarkTheme() ? "rgba(255, 255, 255, 0.1)" : "rgba(0, 0, 0, 0.1)";
}

const formatter = new Intl.NumberFormat("en-US");
const thinSpace = " ";

function formatMoney(value, index, values) {
	return formatter.format(value).replace(",", thinSpace);
};
