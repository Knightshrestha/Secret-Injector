<script lang="ts">
	import type { PageData } from './$types';
	import type { SecretChange, SecretItem } from '$lib/types';
	import { eventEndpoint } from '$lib/url_endpoint';
	import { onMount } from 'svelte';
	import SecretModal from '$lib/components/SecretModal.svelte';
	import DeleteSecretModal from '$lib/components/DeleteSecretModal.svelte';

	let { data }: { data: PageData } = $props();

	let secrets: SecretItem[] = $state(data.secrets || []);
	let isConnected = $state(false);

	let showSecretModal = $state(false);
	let showDeleteModal = $state(false);
	let selectedSecret: SecretItem | null = $state(null);

	onMount(() => {
		const eventSource = new EventSource(eventEndpoint('/secrets'));

		eventSource.onopen = () => {
			console.log('Secret SSE connected');
			isConnected = true;
		};

		eventSource.addEventListener('create', (event) => {
			const change: SecretChange = JSON.parse(event.data);
			secrets = [...secrets, change.data];
		});

		eventSource.addEventListener('update', (event) => {
			const change: SecretChange = JSON.parse(event.data);
			secrets = secrets.map((secret) => (secret.id === change.data.id ? change.data : secret));
		});

		eventSource.addEventListener('delete', (event) => {
			const change: SecretChange = JSON.parse(event.data);
			secrets = secrets.filter((secret) => secret.id !== change.data.id);
		});

		eventSource.addEventListener('ping', () => {});

		eventSource.onerror = (err) => {
			console.error('Secret SSE error:', err);
			isConnected = false;
		};

		return () => {
			eventSource.close();
			isConnected = false;
		};
	});

	function formatDate(dateString: string): string {
		return new Date(dateString).toLocaleString('en-US', {
			year: 'numeric',
			month: 'short',
			day: 'numeric',
			hour: '2-digit',
			minute: '2-digit'
		});
	}

	function openCreateModal() {
		selectedSecret = null;
		showSecretModal = true;
	}

	function openEditModal(secret: SecretItem) {
		selectedSecret = secret;
		showSecretModal = true;
	}

	function openDeleteModal(secret: SecretItem) {
		selectedSecret = secret;
		showDeleteModal = true;
	}

	function handleModalSuccess() {
		// SSE will handle the update automatically
		selectedSecret = null;
	}
</script>

<div class="flex flex-col gap-4 p-4">
	<div class="flex items-center justify-between">
		<h2 class="text-2xl font-bold text-gray-900">Project: {data.project.name}</h2>
		<div class="flex items-center gap-4">
			<div class="flex items-center gap-2 text-xs">
				<span
					class="h-2 w-2 rounded-full {isConnected ? 'bg-green-500' : 'bg-gray-400'} 
					{isConnected ? 'animate-pulse' : ''}"
				></span>
				<span class="text-gray-600">
					{isConnected ? 'Live' : 'Disconnected'}
				</span>
			</div>
			<button
				onclick={openCreateModal}
				class="rounded-lg bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
			>
				+ New Secret
			</button>
		</div>
	</div>

	{#if secrets.length === 0}
		<div class="rounded-lg border-2 border-dashed border-gray-300 bg-gray-50 py-12 text-center">
			<svg
				class="mx-auto h-12 w-12 text-gray-400"
				fill="none"
				stroke="currentColor"
				viewBox="0 0 24 24"
			>
				<path
					stroke-linecap="round"
					stroke-linejoin="round"
					stroke-width="2"
					d="M9 13h6m-3-3v6m5 5H7a2 2 0 01-2-2V5a2 2 0 012-2h5.586a1 1 0 01.707.293l5.414 5.414a1 1 0 01.293.707V19a2 2 0 01-2 2z"
				/>
			</svg>
			<p class="mt-2 text-lg text-gray-500">No secrets yet</p>
			<p class="mt-1 text-sm text-gray-400">Create your first secret to get started</p>
		</div>
	{:else}
		{#each secrets as secret (secret.id)}
			<div
				class="group rounded-xl border border-gray-200 bg-white border-r-4 p-6 shadow-sm hover:border-r-blue-400 transition-colors"
			>
				<div class="flex gap-4">
					<!-- Icon Section (left side) -->
					<div class="shrink-0">
						<div
							class="flex h-12 w-12 items-center justify-center rounded-lg bg-linear-to-br from-blue-500 to-purple-600 text-white"
						>
							<svg class="h-6 w-6" fill="none" stroke="currentColor" viewBox="0 0 24 24">
								<path
									stroke-linecap="round"
									stroke-linejoin="round"
									stroke-width="2"
									d="M3 7v10a2 2 0 002 2h14a2 2 0 002-2V9a2 2 0 00-2-2h-6l-2-2H5a2 2 0 00-2 2z"
								/>
							</svg>
						</div>
					</div>

					<!-- Content Section (right side) -->
					<div class="min-w-0 flex-1">
						<!-- Header: Title/Description and Open button -->
						<header class="mb-4">
								<h3 class="truncate text-xl font-semibold text-gray-900">
									{secret.key}: {secret.value}
								</h3>

								{#if secret.description}
									<p class="mt-1 line-clamp-2 text-sm text-gray-600">
										{secret.description}
									</p>
								{/if}
						</header>

						<!-- Footer: Timestamps and Actions -->
						<footer class="flex items-center justify-between gap-4 border-t border-gray-100 pt-4">
							<div class="flex flex-wrap items-center gap-4 text-xs text-gray-500">
								<span class="flex items-center gap-1.5">
									<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
										/>
									</svg>
									{formatDate(secret.created_at)}
								</span>

								<span class="flex items-center gap-1.5">
									<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
										/>
									</svg>
									{formatDate(secret.updated_at)}
								</span>
							</div>

							<div class="flex shrink-0 gap-2">
								<button
									onclick={() => openEditModal(secret)}
									class="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium text-gray-700 transition-colors hover:bg-gray-100 hover:text-gray-900"
									title="Edit Project"
								>
									<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z"
										/>
									</svg>
									<span>Edit</span>
								</button>

								<button
									onclick={() => openDeleteModal(secret)}
									class="flex items-center gap-1.5 rounded-md px-3 py-1.5 text-sm font-medium text-red-600 transition-colors hover:bg-red-50 hover:text-red-700"
									title="Delete Project"
								>
									<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
										<path
											stroke-linecap="round"
											stroke-linejoin="round"
											stroke-width="2"
											d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
										/>
									</svg>
									<span>Delete</span>
								</button>
							</div>
						</footer>
					</div>
				</div>
			</div>
		
		{/each}
	{/if}
</div>

<SecretModal
	projectId={data.project.id}
	bind:isOpen={showSecretModal}
	secret={selectedSecret}
	onSuccess={handleModalSuccess}
/>
<DeleteSecretModal
	bind:isOpen={showDeleteModal}
	secret={selectedSecret}
	onSuccess={handleModalSuccess}
/>
