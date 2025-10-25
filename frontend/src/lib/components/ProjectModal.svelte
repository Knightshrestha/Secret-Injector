<script lang="ts">
	import type { ProjectItem } from '$lib/types';
	import { apiEndpoint } from '$lib/url_endpoint';

	let {
		isOpen = $bindable(false),
		project = null,
		onSuccess
	}: {
		isOpen: boolean;
		project?: ProjectItem | null;
		onSuccess?: () => void;
	} = $props();

	let name = $state('');
	let description = $state('');
	let isSubmitting = $state(false);
	let error = $state('');

	$effect(() => {
		if (isOpen) {
			if (project) {
				name = project.name;
				description = project.description || '';
			} else {
				name = '';
				description = '';
			}
			error = '';
		}
	});

	function closeModal() {
		isOpen = false;
		error = '';
	}

	async function handleSubmit(e: Event) {
		e.preventDefault();

		if (!name.trim()) {
			error = 'Project name is required';
			return;
		}

		isSubmitting = true;
		error = '';

		try {
			const url = project ? apiEndpoint(`/projects/${project.id}`) : apiEndpoint('/projects');

			const method = project ? 'PATCH' : 'POST';

			const response = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					name: name.trim(),
					description: description.trim() || null
				})
			});

			if (!response.ok) {
				const data = await response.json();
				console.log(data);

				throw new Error(data.error || 'Failed to save project');
			}

			closeModal();
			onSuccess?.();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			isSubmitting = false;
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			closeModal();
		}
	}
</script>

{#if isOpen}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
		onclick={handleBackdropClick}
		tabindex="0"
		onkeydown={(e) => {
			if (e.key === 'Enter' || e.key === ' ') handleBackdropClick;
		}}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
	>
		<div class="w-full max-w-md rounded-lg bg-white shadow-xl">
			<div class="border-b border-gray-200 px-6 py-4">
				<h3 id="modal-title" class="text-xl font-bold text-gray-900">
					{project ? 'Edit Project' : 'Create New Project'}
				</h3>
			</div>

			<form onsubmit={handleSubmit} class="p-6">
				{#if error}
					<div class="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-800">
						{error}
					</div>
				{/if}

				<div class="mb-4">
					<label for="name" class="mb-1 block text-sm font-medium text-gray-700">
						Project Name <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						id="name"
						bind:value={name}
						class="w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
						placeholder="Enter project name"
						required
						disabled={isSubmitting}
					/>
				</div>

				<div class="mb-6">
					<label for="description" class="mb-1 block text-sm font-medium text-gray-700">
						Description
					</label>
					<textarea
						id="description"
						bind:value={description}
						rows="3"
						class="w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
						placeholder="Enter project description (optional)"
						disabled={isSubmitting}
					></textarea>
				</div>

				<div class="flex justify-end gap-3">
					<button
						type="button"
						onclick={closeModal}
						class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
						disabled={isSubmitting}
					>
						Cancel
					</button>
					<button
						type="submit"
						class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:bg-blue-400"
						disabled={isSubmitting}
					>
						{isSubmitting ? 'Saving...' : project ? 'Update' : 'Create'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
