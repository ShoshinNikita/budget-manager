<script lang="ts">
	import type * as types from "@src/types";

	export let account: types.AccountWithBalance;
</script>

<div class="account">
	<div class="name">
		<span>{account.name}</span>
	</div>
	<div class="balance">
		<span>{account.balance}&nbsp;{account.currency}</span>
	</div>
	<div class="actions">
		<button>Edit</button>
		<button>Close</button>
	</div>
</div>

<style>
	.account {
		display: grid;
		grid-template-columns: 1fr 1fr auto;
		padding: 0 5px;
		column-gap: 15px;
		margin: 5px 0;
	}
	.balance {
		text-align: right;
	}
</style>
