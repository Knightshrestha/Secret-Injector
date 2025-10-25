<script lang="ts">
	import { page } from '$app/state';

	const { status, message, details } = $derived({
		status: page.status,
		message: page.error?.message || 'An unexpected error occurred',
		details: (page.error as any)?.details || 'Please try again later'
	});

	const getErrorIcon = (status: number) => {
		if (status === 404) return '🔍';
		if (status === 403) return '🔒';
		if (status >= 500) return '⚠️';
		if (status === 503) return '🌐';
		return '❌';
	};

	const getErrorTitle = (status: number) => {
		if (status === 404) return 'Page Not Found';
		if (status === 403) return 'Access Forbidden';
		if (status >= 500) return 'Server Error';
		if (status === 503) return 'Service Unavailable';
		return 'Error';
	};
</script>

<div class="min-h-screen bg-linear-to-br from-gray-50 to-gray-100 flex items-center justify-center p-4">
	<div class="max-w-md w-full">
		<div class="bg-white rounded-lg shadow-xl p-8 text-center">
			<!-- Error Icon -->
			<div class="text-6xl mb-4">
				{getErrorIcon(status)}
			</div>

			<!-- Status Code -->
			<div class="text-gray-400 text-sm font-semibold uppercase tracking-wider mb-2">
				Error {status}
			</div>

			<!-- Error Title -->
			<h1 class="text-3xl font-bold text-gray-900 mb-3">
				{getErrorTitle(status)}
			</h1>

			<!-- Error Message -->
			<p class="text-lg text-gray-700 mb-2">
				{message}
			</p>

			<!-- Error Details -->
			<p class="text-sm text-gray-500 mb-8">
				{details}
			</p>

			<!-- Action Buttons -->
			<div class="flex flex-col sm:flex-row gap-3 justify-center">
				<a
					href="/"
					class="px-6 py-3 bg-blue-600 text-white rounded-lg font-medium hover:bg-blue-700 transition-colors"
				>
					Go Home
				</a>
				<button
					onclick={() => window.history.back()}
					class="px-6 py-3 bg-gray-200 text-gray-700 rounded-lg font-medium hover:bg-gray-300 transition-colors"
				>
					Go Back
				</button>
			</div>
		</div>

		<!-- Additional Help Text -->
		<p class="text-center text-gray-600 text-sm mt-6">
			If this problem persists, please contact support.
		</p>
	</div>
</div>