<script lang="ts">
	import * as types from "@src/types";

	export let account: types.AccountWithBalance;

	const isAccountOpen = account.status === types.AccountStatus.Open;
</script>

<div class="account">
	<div class="name">
		<span>{account.name}</span>
	</div>
	<div class="balance">
		<span>{account.balance}&nbsp;{account.currency}</span>
	</div>
	<div class="actions">
		<button disabled={!isAccountOpen}>Edit</button>
		<button disabled={!isAccountOpen}>Close</button>
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
