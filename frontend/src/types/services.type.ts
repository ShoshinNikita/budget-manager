import type { AccountWithBalance } from "./account.type";
import type { Currency } from "./currency.type";

export interface NotificationService {
	notify(msg: string): void;
}

export interface AccountService {
	getOpenAccounts(f: (_: AccountWithBalance[]) => void): void;
	getClosedAccounts(f: (_: AccountWithBalance[]) => void): void;

	createAccount(name: string, currency: Currency): Promise<boolean>;
	editAccount(id: string, newName?: string): Promise<void>;
	closeAccount(acc: AccountWithBalance): Promise<void>;
}

export interface TransactionService {
	createTransferTransaction(args: CreateTransferTransactionArgs): Promise<boolean>;
}

export type CreateTransferTransactionArgs = {
	fromAccountID: string;
	fromAmount: number;

	toAccountID: string;
	toAmount: number;
};
