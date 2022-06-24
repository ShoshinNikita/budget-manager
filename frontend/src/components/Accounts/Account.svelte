<script lang="ts">
	import * as types from "@src/types";
	import Button from "@src/components/Button.svelte";

	export let account: types.AccountWithBalance;

	const isAccountOpen = account.status === types.AccountStatus.Open;
</script>

<div class="account" class:closed={!isAccountOpen}>
	<div class="name">
		<span>{account.name}</span>
	</div>
	<div class="balance">
		<span>{account.balance}&nbsp;{account.currency}</span>
	</div>
	<div class="actions">
		<Button icon="edit" disabled={!isAccountOpen} title="Edit" />
		<Button icon="x" disabled={!isAccountOpen} title="Close" />
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

	.closed > .name {
		text-decoration: line-through;
	}

	.balance {
		text-align: right;
	}
</style>
