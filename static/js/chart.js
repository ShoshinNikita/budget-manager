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
	global.tooltips.bodyFontSize = 14;
	global.tooltips.backgroundColor = "#000000d0";
	global.tooltips.cornerRadius = 5;
	// Other
	global.maintainAspectRatio = false;

	// Tune scale
	const scale = Chart.defaults.scale;
	scale.ticks.fontSize = 14;
	scale.beginAtZero = true;
})();
