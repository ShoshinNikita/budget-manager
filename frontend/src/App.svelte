<script>
	import "@src/styles/card.scss";
	import "@src/styles/global.scss";
	import "@src/styles/inputs.scss";
	import "@src/styles/scrollbar.scss";

	import { Router, Route } from "svelte-routing";
	import Main from "@src/routes/Main.svelte";
	import Statistics from "@src/routes/Statistics.svelte";

	export let url = "";
</script>

<Router {url}>
	<div id="router">
		<Route path="/statistics" component={Statistics} />
		<Route path="/" component={Main} />
	</div>
</Router>

<style lang="scss">
	#router {
		height: 100%;
	}
</style>
