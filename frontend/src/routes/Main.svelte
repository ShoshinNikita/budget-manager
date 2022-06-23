<script lang="ts">
	import Accounts from "@src/components/Accounts";
	import Calendar from "@src/components/Calendar";
	import * as types from "@src/types";

	const accounts: types.AccountWithBalance[] = [
		{
			id: "123",
			name: "USD account",
			balance: "15.44",
			currency: "USD",
			status: types.AccountStatus.Open,
		} as types.AccountWithBalance,
		{
			id: "456",
			name: "RUB account",
			balance: "1555",
			currency: "RUB",
			status: types.AccountStatus.Open,
		} as types.AccountWithBalance,
		{
			id: "789",
			name: "RUB account",
			balance: "1555",
			currency: "RUB",
			status: types.AccountStatus.Close,
		} as types.AccountWithBalance,
	];
</script>

<div class="main">
	<Accounts {accounts} />
	<Calendar />
</div>

<style lang="scss">
	.main {
		display: grid;
		grid-template-columns: 2fr 6fr;
		column-gap: 15px;
		height: 100%;
	}
</style>
