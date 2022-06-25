<script lang="ts">
	import Accounts from "@src/components/Accounts";
	import Calendar from "@src/components/Calendar";
</script>

<div class="main">
	<Accounts />
	<Calendar />
</div>

<style lang="scss">
	.main {
		display: grid;
		grid-template-columns: 2fr 7fr;
		column-gap: 15px;
		height: 100%;
	}
</style>
