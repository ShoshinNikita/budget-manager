<script lang="ts" context="module">
	export enum ButtonSize {
		Small = 0,
		Medium = 1,
		Large = 2,
	}
</script>

<script lang="ts">
	import feather from "feather-icons";

	export let icon = "";
	export let size: ButtonSize = ButtonSize.Small;
	export let disabled = false;
	export let title = "";
	export let onClick = () => {};

	let sizeInPx = "20px";
	switch (size) {
		case ButtonSize.Medium:
			sizeInPx = "30px";
			break;
		case ButtonSize.Large:
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

<button {title} on:click={onClick} {disabled}>
	{@html svgIcon}
</button>

<style lang="scss">
	button {
		background-color: #00000000;
		border: none;
		cursor: pointer;
		opacity: 0.6;
		padding: 0;
	}

	button:hover,
	button:focus {
		&:not(:disabled) {
			opacity: 1;
		}
	}
</style>
