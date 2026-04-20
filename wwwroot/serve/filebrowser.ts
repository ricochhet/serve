const BASE: string = window.location.pathname.replace(/\/$/, "");
let currentPath: string = "/";

const joinPath = (base: string, name: string): string =>
	(base === "/" ? "" : base) + "/" + name;

const formatSize = (bytes: number): string => {
	if (bytes === 0) return "None";
	const units = ["B", "KB", "MB", "GB", "TB"] as const;
	let i = 0;

	while (bytes >= 1024 && i < units.length - 1) {
		bytes /= 1024;
		i++;
	}

	return (i === 0 ? bytes : bytes.toFixed(1)) + "\u202f" + units[i];
};

const formatDate = (value: string | number | Date): string => {
	const d = new Date(value);

	return `${d.toLocaleDateString(undefined, {
		year: "numeric",
		month: "short",
		day: "numeric",
	})} ${d.toLocaleTimeString(undefined, {
		hour: "2-digit",
		minute: "2-digit",
	})}`;
};

const parentOf = (path: string): string => {
	if (path === "/") return "/";
	const trimmed = path.replace(/\/$/, "");
	const idx = trimmed.lastIndexOf("/");
	return idx <= 0 ? "/" : trimmed.slice(0, idx);
};

interface Entry {
	name: string;
	isDir: boolean;
	size?: number;
	modTime?: string | number | Date;
	_nav?: string;
	_parentRow?: boolean;
}

interface ApiResponse {
	path: string;
	entries?: Entry[];
}

function renderBreadcrumb(path: string): void {
	const el = document.getElementById("breadcrumb") as HTMLElement;
	const parts = path.split("/").filter(Boolean);
	let acc = "";

	const html = [
		`<a href="#" data-nav="/">~</a>`,
		...parts.flatMap((p, i) => {
			acc += "/" + p;
			const nav = acc;

			return [
				`<span class="sep">/</span>`,
				i === parts.length - 1
					? `<span class="current">${p}</span>`
					: `<a href="#" data-nav="${nav}">${p}</a>`,
			];
		}),
	].join("");

	el.innerHTML = html;

	el.querySelectorAll<HTMLAnchorElement>("a[data-nav]").forEach((a) => {
		a.onclick = (e: MouseEvent) => {
			e.preventDefault();
			navigate(a.dataset.nav!);
		};
	});
}

async function navigate(path: string): Promise<void> {
	currentPath = path;

	const err = document.getElementById("error-msg") as HTMLElement;
	err.classList.add("slv-hidden");

	let data: ApiResponse;

	try {
		const res = await fetch(`${BASE}/api?path=${encodeURIComponent(path)}`);
		if (!res.ok) throw new Error(await res.text());
		data = (await res.json()) as ApiResponse;
	} catch (e: any) {
		err.textContent = "Error: " + e.message;
		err.classList.remove("slv-hidden");
		return;
	}

	renderBreadcrumb(data.path);

	const tbody = document.getElementById("entries") as HTMLElement;
	tbody.innerHTML = "";

	if (data.path !== "/") {
		tbody.appendChild(
			makeRow(
				{
					name: "../",
					isDir: true,
					_nav: parentOf(data.path),
					_parentRow: true,
				},
				data.path,
			),
		);
	}

	(data.entries || []).forEach((e) => {
		tbody.appendChild(makeRow(e, data.path));
	});

	if (!data.entries?.length) {
		const tr = document.createElement("tr");
		tr.innerHTML = `<td colspan="3" style="color:rgba(255,255,255,0.3);padding:14px 8px">empty directory</td>`;
		tbody.appendChild(tr);
	}
}

function makeRow(e: Entry, path: string): HTMLTableRowElement {
	const tr = document.createElement("tr");

	if (e.isDir) {
		const nav = e._parentRow ? e._nav! : joinPath(path, e.name);
		const label = e._parentRow ? "../" : e.name + "/";

		tr.innerHTML = `
			<td>
				<div class="entry-name">
					<span class="entry-icon dir">▶</span>
					<a href="#" data-nav="${nav}">${label}</a>
				</div>
			</td>
			<td class="col-size">None</td>
			<td class="col-date">${e.modTime ? formatDate(e.modTime) : ""}</td>
		`;

		tr.querySelector<HTMLAnchorElement>("a")!.onclick = (
			ev: MouseEvent,
		) => {
			ev.preventDefault();
			navigate(nav);
		};
	} else {
		tr.innerHTML = `
			<td>
				<div class="entry-name">
					<span class="entry-icon">·</span>
					<span>${e.name}</span>
				</div>
			</td>
			<td class="col-size">${formatSize(e.size ?? 0)}</td>
			<td class="col-date">${formatDate(e.modTime!)}</td>
		`;
	}

	return tr;
}

navigate("/");
