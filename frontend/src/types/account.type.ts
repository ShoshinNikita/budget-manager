import type { Currency } from "./currency.type";

export type Account = {
	id: string;
	name: string;
	currency: Currency;
	status: AccountStatus;
	createdAt: Date;
	updatedAt: Date;
};

export enum AccountStatus {
	Open = "open",
	Close = "close",
}

export type AccountWithBalance = Account & {
	balance: string;
};
