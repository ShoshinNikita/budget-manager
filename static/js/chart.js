// Update global Chart.js options
(function () {
	const global = Chart.defaults.global;
	// Hide legend
	global.legend.display = false;
	// Disable animations
	global.animation.duration = 0;
	global.hover.animationDuration = 0;
	global.responsiveAnimationDuration = 0;
	// Tune tooltips
	global.tooltips.titleFontSize = 15;
	global.tooltips.backgroundColor = "#000000d0";
	global.tooltips.cornerRadius = 5;
	// Other
	global.defaultFontSize = 14;
	global.maintainAspectRatio = false;

	// Tune scale
	const scale = Chart.defaults.scale;
	scale.ticks.beginAtZero = true;
	scale.gridLines.color = getGridLinesColor();
})();

function getGridLinesColor() {
	return isDarkTheme() ? "rgba(255, 255, 255, 0.1)" : "rgba(0, 0, 0, 0.1)";
}
