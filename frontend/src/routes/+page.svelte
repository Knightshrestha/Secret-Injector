<script lang="ts">
	import type { PageData } from './$types';
	import type { ProjectChange, ProjectItem } from '$lib/types';
	import { eventEndpoint } from '$lib/url_endpoint';
	import { onMount } from 'svelte';
	import ProjectModal from '$lib/components/ProjectModal.svelte';
	import DeleteProjectModal from '$lib/components/DeleteProjectModal.svelte';

	let { data }: { data: PageData } = $props();

	let projects: ProjectItem[] = $state(data.projects || []);
	let isConnected = $state(false);

	let showProjectModal = $state(false);
	let showDeleteModal = $state(false);
	let selectedProject: ProjectItem | null = $state(null);

	onMount(() => {
		const eventSource = new EventSource(eventEndpoint('/projects'));

		eventSource.onopen = () => {
			console.log('Project SSE connected');
			isConnected = true;
		};

		eventSource.addEventListener('create', (event) => {
			const change: ProjectChange = JSON.parse(event.data);
			projects = [...projects, change.data];
		});

		eventSource.addEventListener('update', (event) => {
			const change: ProjectChange = JSON.parse(event.data);
			projects = projects.map((project) => (project.id === change.data.id ? change.data : project));
		});

		eventSource.addEventListener('delete', (event) => {
			const change: ProjectChange = JSON.parse(event.data);
			projects = projects.filter((project) => project.id !== change.data.id);
		});

		eventSource.addEventListener('ping', () => {});

		eventSource.onerror = (err) => {
			console.error('Project SSE error:', err);
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
		selectedProject = null;
		showProjectModal = true;
	}

	function openEditModal(project: ProjectItem) {
		selectedProject = project;
		showProjectModal = true;
	}

	function openDeleteModal(project: ProjectItem) {
		selectedProject = project;
		showDeleteModal = true;
	}

	function handleModalSuccess() {
		// SSE will handle the update automatically
		selectedProject = null;
	}
</script>

<div class="flex flex-col gap-4 p-4">
	<div class="flex items-center justify-between">
		<h2 class="text-2xl font-bold text-gray-900">Projects</h2>
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
				+ New Project
			</button>
		</div>
	</div>

	{#if projects.length === 0}
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
			<p class="mt-2 text-lg text-gray-500">No projects yet</p>
			<p class="mt-1 text-sm text-gray-400">Create your first project to get started</p>
		</div>
	{:else}
		{#each projects as project (project.id)}
			<div
				class="group rounded-xl border border-gray-100 bg-white p-6 shadow-sm transition-all hover:shadow-md"
			>
				<div class="mb-3 flex items-start justify-between">
					<div class="flex-1">
						<h3 class="mb-1 text-xl font-bold text-gray-900">
							{project.name}
						</h3>

						{#if project.description}
							<p class="text-sm text-gray-600">
								{project.description}
							</p>
						{/if}
					</div>

					<div class="flex gap-2 opacity-0 transition-opacity group-hover:opacity-100">
						<button
							onclick={() => openEditModal(project)}
							class="text-sm font-medium text-blue-600 hover:text-blue-800"
						>
							Edit
						</button>
						<button
							onclick={() => openDeleteModal(project)}
							class="text-sm font-medium text-red-600 hover:text-red-800"
						>
							Delete
						</button>
					</div>
				</div>

				<a href={`projects/${project.id}`}>Open Project</a>

				<div class="flex items-center gap-6 border-t border-gray-100 pt-3 text-xs text-gray-500">
					<span class="flex items-center gap-1">
						<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"
							/>
						</svg>
						Created {formatDate(project.created_at)}
					</span>
					<span class="flex items-center gap-1">
						<svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
							<path
								stroke-linecap="round"
								stroke-linejoin="round"
								stroke-width="2"
								d="M12 8v4l3 3m6-3a9 9 0 11-18 0 9 9 0 0118 0z"
							/>
						</svg>
						Updated {formatDate(project.updated_at)}
					</span>
				</div>
			</div>
		{/each}
	{/if}
</div>

<ProjectModal
	bind:isOpen={showProjectModal}
	project={selectedProject}
	onSuccess={handleModalSuccess}
/>
<DeleteProjectModal
	bind:isOpen={showDeleteModal}
	project={selectedProject}
	onSuccess={handleModalSuccess}
/>
