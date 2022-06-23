import type * as types from "@src/types";

export const getAccounts = async (): Promise<types.AccountWithBalance[] | void> => {
	const resp = await makeRequest("/api/accounts/get", "Couldn't get accounts");
	if (!resp) {
		return;
	}

	let res: types.AccountWithBalance[] = [];
	for (let account of resp["accounts"]) {
		res.push({
			id: account["id"],
			name: account["name"],
			balance: account["balance"],
			currency: account["currency"],
			status: account["status"],
			createdAt: new Date(account["created_at"]),
			updatedAt: new Date(account["updated_at"]),
		});
	}

	return res;
};

const backendURL = import.meta.env.VITE_BACKEND_API_URL;

const makeRequest = async (path: string, errMsg: string): Promise<object | void> => {
	const startTime = new Date();
	const logRequestTime = () => {
		const since = new Date().getTime() - startTime.getTime();
		console.log(`request "POST ${path}" took ${since} ms`);
	};

	return fetch(backendURL + path, { method: "POST" })
		.then(async (resp) => {
			const respText = await resp.text();

			if (resp.status !== 200) {
				logRequestTime();
				handleError(errMsg, `got unexpected status code ${resp.status}, body: "${respText}"`);
				return;
			}

			const parsedResp = JSON.parse(respText);
			if (typeof parsedResp !== "object") {
				logRequestTime();
				handleError(errMsg, `got unexpected response type: ${typeof parsedResp}`);
				return;
			}

			logRequestTime();
			return parsedResp;
		})
		.catch((err) => {
			logRequestTime();
			handleError(errMsg, err);
		});
};

const handleError = (errMsg: string, err: any) => {
	// TODO: show notification
	console.log(`${errMsg}: ${err}`);
};
