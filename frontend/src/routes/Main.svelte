<script lang="ts">
	import Accounts from "@src/components/Accounts";
	import Transfer from "@src/components/Transfer";
	import Calendar from "@src/components/Calendar";
</script>

<div class="main">
	<span id="accounts" class="wrapper">
		<Accounts />
	</span>
	<span id="transfer" class="wrapper">
		<Transfer />
	</span>
	<span id="calendar" class="wrapper">
		<Calendar />
	</span>
</div>

<style lang="scss">
	$gap: 15px;

	.main {
		display: grid;
		grid-template-columns: 2fr 7fr;
		grid-template-rows: auto min-content;
		column-gap: $gap;
		row-gap: $gap;
		height: 100%;
	}

	.wrapper {
		overflow: auto;
	}

	#accounts {
		grid-column: 1 / 2;
		grid-row: 1 / 2;
	}

	#transfer {
		grid-column: 1 / 2;
		grid-row: 2 / 3;
	}

	#calendar {
		grid-column: 2 / 3;
		grid-row: 1 / 3;
	}
</style>
