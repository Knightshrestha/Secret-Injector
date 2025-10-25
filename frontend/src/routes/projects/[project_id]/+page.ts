import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';
import { apiEndpoint } from '$lib/url_endpoint';
import type { ProjectItem, SecretItem } from '$lib/types';

export const load = (async ({ params, fetch }) => {
	if (!params.project_id) {
		throw error(404, {
			message: 'Project ID not provided'
			// details: 'No project_id parameter was provided'
		});
	}

	const projectRequest = await fetch(apiEndpoint(`/projects/${params.project_id}`));

	if (projectRequest.ok) {
		const project: ProjectItem = await projectRequest.json();

		const secretRequest = await fetch(apiEndpoint(`/projects/${project.id}/secrets`));

		if (secretRequest.ok) {
			const secrets: SecretItem[] = await secretRequest.json();
			return {
				secrets,
				project
			};
		} else {
			throw error(secretRequest.status);
		}
	} else {
		throw error(projectRequest.status, await projectRequest.json());
	}
}) satisfies PageLoad;
