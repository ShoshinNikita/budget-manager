<script>
	import { Router, Route } from "svelte-routing";
	import Main from "@src/routes/Main.svelte";
	import Statistics from "@src/routes/Statistics.svelte";

	export let url = "";
</script>

<Router {url}>
	<div>
		<Route path="/statistics" component={Statistics} />
		<Route path="/" component={Main} />
	</div>
</Router>
