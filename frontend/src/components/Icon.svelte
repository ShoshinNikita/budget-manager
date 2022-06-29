<script lang="ts" context="module">
	export enum Size {
		Small = 0,
		Medium = 1,
		Large = 2,
	}
</script>

<script lang="ts">
	import feather from "feather-icons";

	export let icon: string;
	export let size: Size = Size.Small;

	let sizeInPx = "20px";
	switch (size) {
		case Size.Medium:
			sizeInPx = "30px";
			break;
		case Size.Large:
			sizeInPx = "40px";
			break;
	}

	let svgIcon = "";
	switch (icon) {
		// Custom icons
		case "transfer":
			// Source: https://github.com/feathericons/feather/pull/580
			svgIcon = `
				<svg
					width="${sizeInPx}" height="${sizeInPx}" viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg"
					fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"
				>
					<path d="M14.5 4l4 4-4 4" />
					<path d="M9 20l-4-4 4-4" />
					<path d="M3 8h15" />
					<path d="M6 16h15" />
				</svg>`;
			break;

		default:
			svgIcon = feather.icons[icon].toSvg({
				width: sizeInPx,
				height: sizeInPx,
			});
	}
</script>

<span>
	{@html svgIcon}
</span>
