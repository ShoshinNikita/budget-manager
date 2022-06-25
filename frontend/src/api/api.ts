export class Error {
	cause: string;

	constructor(cause: string) {
		this.cause = cause;
	}
}

const backendURL = import.meta.env.VITE_BACKEND_API_URL;

export async function makeRequest<T>(path: string, req: object): Promise<Error | T> {
	const startTime = new Date();
	const logRequestTime = () => {
		const since = new Date().getTime() - startTime.getTime();
		console.log(`request "POST ${path}" took ${since} ms`);
	};

	return fetch(backendURL + path, {
		method: "POST",
		body: JSON.stringify(req),
	})
		.then(async (resp) => {
			if (resp.ok) {
				try {
					return (await resp.json()) as T;
				} catch (err) {
					return new Error(`couldn't parse response: ${err}`);
				}
			}

			try {
				const body = await resp.text();
				return new Error(`got unexpected status code ${resp.status}, body: "${body}"`);
			} catch (err) {
				return new Error(`couldn't read body with error: ${err}`);
			}
		})
		.then((res) => {
			logRequestTime();
			return res;
		})
		.catch((err) => {
			logRequestTime();
			return new Error(err);
		});
}
