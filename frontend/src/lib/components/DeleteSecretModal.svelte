<script lang="ts">
	import type { SecretItem } from '$lib/types';
	import { apiEndpoint } from '$lib/url_endpoint';

	let {
		isOpen = $bindable(false),
		secret,
		onSuccess
	}: {
		isOpen: boolean;
		secret: SecretItem | null;
		onSuccess?: () => void;
	} = $props();

	let isDeleting = $state(false);
	let error = $state('');

	$effect(() => {
		if (isOpen) {
			error = '';
		}
	});

	function closeModal() {
		isOpen = false;
		error = '';
	}

	async function handleDelete() {
		if (!secret) return;

		isDeleting = true;
		error = '';

		try {
			const response = await fetch(apiEndpoint(`/secrets/${secret.id}`), {
				method: 'DELETE'
			});

			if (!response.ok) {
				const data = await response.json();
				throw new Error(data.message || 'Failed to delete secret');
			}

			closeModal();
			onSuccess?.();
		} catch (err) {
			error = err instanceof Error ? err.message : 'An error occurred';
		} finally {
			isDeleting = false;
		}
	}

	function handleBackdropClick(e: MouseEvent) {
		if (e.target === e.currentTarget) {
			closeModal();
		}
	}
</script>

{#if isOpen && secret}
	<div
		class="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4"
		onclick={handleBackdropClick}
		role="dialog"
		aria-modal="true"
		aria-labelledby="modal-title"
		tabindex="0"
		onkeydown={(e) => {
			// if (e.key === 'Enter' || e.key === ' ') handleBackdropClick;
		}}
	>
		<div class="w-full max-w-md rounded-lg bg-white shadow-xl">
			<div class="border-b border-gray-200 px-6 py-4">
				<h3 id="modal-title" class="text-xl font-bold text-gray-900">Delete Secret</h3>
			</div>

			<div class="p-6">
				{#if error}
					<div class="mb-4 rounded-lg bg-red-50 p-3 text-sm text-red-800">
						{error}
					</div>
				{/if}

				<p class="mb-2 text-gray-700">
					Are you sure you want to delete <strong>{secret.key}</strong>?
				</p>
				<p class="text-sm text-gray-500">
					This action cannot be undone. This will also delete all associated secrets.
				</p>
			</div>

			<div class="flex justify-end gap-3 border-t border-gray-200 px-6 py-4">
				<button
					type="button"
					onclick={closeModal}
					class="rounded-lg border border-gray-300 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-50"
					disabled={isDeleting}
				>
					Cancel
				</button>
				<button
					type="button"
					onclick={handleDelete}
					class="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-700 disabled:bg-red-400"
					disabled={isDeleting}
				>
					{isDeleting ? 'Deleting...' : 'Delete'}
				</button>
			</div>
		</div>
	</div>
{/if}
