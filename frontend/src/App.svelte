<script>
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

<style global lang="scss">
	* {
		box-sizing: border-box;
		font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto,
			Oxygen-Sans, Ubuntu, Cantarell, "Helvetica Neue", sans-serif;
		font-weight: normal;
	}

	body {
		height: 100vh;
		margin: 0;
		padding: 0;
		width: 100vw;
	}

	#app {
		margin: 0;
		padding: 10px;
		height: 100%;
	}

	#router {
		height: 100%;
	}

	.card {
		border: 1px solid black;
		padding: 10px;
	}

	h2.card-title {
		border-bottom: 1px solid rgb(211, 211, 211);
		margin: 0 0 10px;
		padding-left: 5px;
		text-align: left;
	}

	button {
		cursor: pointer;
	}
</style>
