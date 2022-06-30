import * as types from "@src/types";
import * as api from "@src/api";

export class AccountService {
	store: types.AccountStore;
	notificationService: types.NotificationService;

	constructor(store: types.AccountStore, notificationService: types.NotificationService) {
		this.store = store;
		this.notificationService = notificationService;

		this.refreshStore();
	}

	getOpenAccounts = (f: (_: types.AccountWithBalance[]) => void) => {
		this.store.openAccountStore.subscribe(f);
	};

	getClosedAccounts = (f: (_: types.AccountWithBalance[]) => void) => {
		this.store.closedAccountStore.subscribe(f);
	};

	createAccount = async (name: string, currency: types.Currency): Promise<boolean> => {
		const resp = await api.createAccount(name, currency);
		if (resp instanceof api.Error) {
			this.notificationService.notify(`couldn't create account: ${resp.cause}`);
			return false;
		}

		await this.refreshStore();

		return true;
	};

	editAccount = async (id: string, newName?: string) => {
		this.notificationService.notify("account editing is not implemented yet");
	};

	closeAccount = async (account: types.AccountWithBalance) => {
		if (!confirm(`Do you really want to close account "${account.name}"`)) {
			return;
		}

		const resp = await api.closeAccount(account.id);
		if (resp instanceof api.Error) {
			this.notificationService.notify(`couldn't close account: ${resp.cause}`);
			return;
		}

		await this.refreshStore();
	};

	refreshStore = async () => {
		const resp = await api.getAccounts();
		if (resp instanceof api.Error) {
			this.notificationService.notify(`couldn't refresh accounts: ${resp.cause}`);
			return;
		}

		let openAccounts: types.AccountWithBalance[] = [];
		let closedAccounts: types.AccountWithBalance[] = [];
		for (const acc of resp) {
			if (acc.status == types.AccountStatus.Open) {
				openAccounts.push(acc);
			} else {
				closedAccounts.push(acc);
			}
		}

		this.store.openAccountStore.set(openAccounts);
		this.store.closedAccountStore.set(closedAccounts);
	};
}
