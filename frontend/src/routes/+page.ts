import type { ProjectItem } from '$lib/types';
import { apiEndpoint } from '$lib/url_endpoint';
import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';

class CustomError extends Error {
	details: string;
	constructor(message: string, details: string) {
		super(message);
		this.details = details;
	}
}

export const load = (async ({ fetch }) => {
	try {
		const response = await fetch(apiEndpoint('/projects'));

		if (!response.ok) {
			// Throw appropriate errors based on status code
			if (response.status === 404) {
				throw error(
					404,
					new CustomError(
						'Projects endpoint not found',
						'The server could not find the projects resource'
					)
				);
			}

			if (response.status === 401 || response.status === 403) {
				throw error(403);
			}

			if (response.status === 500) {
				throw error(
					500,
					new CustomError('Server error', 'The server encountered an error while fetching projects')
				);
			}

			// Generic error for other status codes
			throw error(
				response.status,
				new CustomError(
					'Failed to fetch projects',
					`Server returned status ${response.status}: ${response.statusText}`
				)
			);
		}

		const projects: ProjectItem[] = await response.json();

		return { projects };
	} catch (err) {
		// If it's already a SvelteKit error, re-throw it
		if (err && typeof err === 'object' && 'status' in err) {
			throw err;
		}

		// Network errors or JSON parse errors
		console.error('Error loading projects:', err);
		throw error(
			503,
			new CustomError(
				'Failed to connect to server',
				'Please check your network connection and try again'
			)
		);
	}
}) satisfies PageLoad;
