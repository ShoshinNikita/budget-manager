import type * as types from "@src/types";
import { NotificationService } from "./notifications";
import { AccountService } from "./accounts";
import { AccountStore } from "@src/stores/accounts";
import { TransactionService } from "./transactions";

const accountStore = new AccountStore();

export const notificationService: types.NotificationService = new NotificationService();
export const accountService: types.AccountService = new AccountService(accountStore, notificationService);
export const transactionService: types.TransactionService = new TransactionService(notificationService);
