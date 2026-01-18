export type DataRow = Record<string, string | number | boolean | null>;

export interface Column {
	name: string;
	dataType: string;
	isEnriching?: boolean;
}
