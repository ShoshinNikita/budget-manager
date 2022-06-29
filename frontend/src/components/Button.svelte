<script lang="ts" context="module">
	export { Size } from "./Icon.svelte";
</script>

<script lang="ts">
	import Icon, { Size } from "./Icon.svelte";

	export let icon = "";
	export let size: Size = Size.Small;
	export let disabled = false;
	export let title = "";
	export let onClick = () => {};
</script>

<button {title} on:click={onClick} {disabled}>
	<Icon {icon} {size} />
</button>

<style lang="scss">
	button {
		background-color: #00000000;
		border: none;
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
