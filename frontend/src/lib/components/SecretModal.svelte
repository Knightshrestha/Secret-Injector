<script lang="ts">
	import type { SecretItem } from '$lib/types';
	import { apiEndpoint } from '$lib/url_endpoint';

	let {
		projectId,
		isOpen = $bindable(false),
		secret = null,
		onSuccess
	}: {
		projectId: string,
		isOpen: boolean;
		secret?: SecretItem | null;
		onSuccess?: () => void;
	} = $props();

	let key = $state('');
	let description = $state('');
	let value = $state('');
	let isSubmitting = $state(false);
	let error = $state('');

	$effect(() => {
		if (isOpen) {
			if (secret) {
				key = secret.key;
				description = secret.description || '';
				value = secret.value || '';
			} else {
				key = '';
				description = '';
				value = '';
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

		if (!key.trim()) {
			error = 'Secret Key is required';
			return;
		}

		if (!value.trim()) {
			error = 'Secret Value is required';
			return;
		}

		isSubmitting = true;
		error = '';

		try {
			const url = secret ? apiEndpoint(`/secrets/${secret.id}`) : apiEndpoint('/secrets');

			const method = secret ? 'PATCH' : 'POST';

			const response = await fetch(url, {
				method,
				headers: {
					'Content-Type': 'application/json'
				},
				body: JSON.stringify({
					project_id: projectId,
					key: key.trim(),
					value: value.trim(),
					description: description.trim() || null
				})
			});

			if (!response.ok) {
				const data = await response.json();
				console.log(data);

				throw new Error(data.error || 'Failed to save secret');
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
					{secret ? 'Edit Secret' : 'Create New Secret'}
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
						Secret Key <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						id="name"
						bind:value={key}
						class="w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
						placeholder="Enter secret key"
						required
						disabled={isSubmitting}
					/>
				</div>

						<div class="mb-4">
					<label for="name" class="mb-1 block text-sm font-medium text-gray-700">
						Secret Value <span class="text-red-500">*</span>
					</label>
					<input
						type="text"
						id="name"
						bind:value={value}
						class="w-full rounded-lg border border-gray-300 px-3 py-2 focus:border-blue-500 focus:ring-1 focus:ring-blue-500 focus:outline-none"
						placeholder="Enter secret value"
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
						placeholder="Enter secret description (optional)"
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
						{isSubmitting ? 'Saving...' : secret ? 'Update' : 'Create'}
					</button>
				</div>
			</form>
		</div>
	</div>
{/if}
