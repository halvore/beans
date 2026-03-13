import { uniqueNamesGenerator, adjectives, animals } from 'unique-names-generator';

function randomSuffix(length = 4): string {
	const chars = 'abcdefghijklmnopqrstuvwxyz0123456789';
	let result = '';
	for (let i = 0; i < length; i++) {
		result += chars[Math.floor(Math.random() * chars.length)];
	}
	return result;
}

export function generateWorkspaceName(): string {
	const base = uniqueNamesGenerator({
		dictionaries: [adjectives, animals],
		separator: '-',
		length: 2
	});
	return `${base}-${randomSuffix()}`;
}
