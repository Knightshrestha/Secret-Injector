export interface ProjectItem {
	id: string;
	name: string;
	description: null | string;
	created_at: string; // ISO string
	updated_at: string;
}

export interface SecretItem {
	id: string;
	project_id: string;
	description: null | string;
	key: string;
	value: string;
	created_at: string;
	updated_at: string;
}

export interface SSE_CHANGE<T> {
	type: 'create' | 'update' | 'delete' | 'ping';
	timestamp: string;
	data: T;
}

export type ProjectChange = SSE_CHANGE<ProjectItem>;
export type SecretChange = SSE_CHANGE<SecretItem>;
