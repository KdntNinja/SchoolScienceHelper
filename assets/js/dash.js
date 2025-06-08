// dash.js - handles dashboard project and activity rendering
fetch('/api/project/list', { credentials: 'include' })
	.then(r => r.ok ? r.json() : [])
	.then(projects => {
		const list = document.getElementById('project-list');
		const empty = document.getElementById('project-list-empty');
		if (empty) empty.remove();
		if (!projects || projects.length === 0) {
			list.innerHTML = '<li class="text-muted-foreground text-center">No projects yet. Click "+ New Project" to get started!</li>';
			return;
		}
		list.innerHTML = '';
		for (const p of projects) {
			const href = p.public_id ? `/project/${p.public_id}` : '#';
			const li = document.createElement('li');
			li.className = 'bg-muted rounded-md p-4 flex items-center justify-between';
			li.innerHTML = `<span class='font-medium'></span><a class='text-primary hover:underline text-sm' href='${href}'>Open</a>`;
			li.querySelector('span').textContent = p.name;
			list.appendChild(li);
		}

		// Render recent activity (sorted by updated_at desc)
		const activityList = document.getElementById('recent-activity-list');
		const activityEmpty = document.getElementById('recent-activity-empty');
		if (activityEmpty) activityEmpty.remove();
		if (!projects || projects.length === 0) {
			activityList.innerHTML = '<li class="text-muted-foreground text-center">No recent activity.</li>';
			return;
		}
		// Sort by updated_at descending
		projects.sort((a, b) => new Date(b.updated_at) - new Date(a.updated_at));
		for (const p of projects.slice(0, 5)) {
			const date = new Date(p.updated_at);
			const ago = timeAgo(date);
			const li = document.createElement('li');
			li.innerHTML = `Edited <span class='font-medium text-primary'></span> ${ago}`;
			li.querySelector('span').textContent = p.name;
			activityList.appendChild(li);
		}
	})
	.catch(() => {
		const list = document.getElementById('project-list');
		list.innerHTML = '<li class="text-muted-foreground text-center">Failed to load projects.</li>';
		const activityList = document.getElementById('recent-activity-list');
		activityList.innerHTML = '<li class="text-muted-foreground text-center">Failed to load activity.</li>';
	});

// New project button handler
document.getElementById('new-project-btn').addEventListener('click', async () => {
	const name = prompt('Enter a name for your new project:');
	if (!name) return;
	const res = await fetch('/api/project/save', {
		method: 'POST',
		headers: { 'Content-Type': 'application/json' },
		credentials: 'include',
		body: JSON.stringify({ name, data: {} })
	});
	if (res.ok) {
		location.reload();
	} else {
		alert('Failed to create project.');
	}
});

// Helper: time ago formatting
function timeAgo(date) {
	const now = new Date();
	const seconds = Math.floor((now - date) / 1000);
	if (seconds < 60) return `${seconds}s ago`;
	const minutes = Math.floor(seconds / 60);
	if (minutes < 60) return `${minutes}m ago`;
	const hours = Math.floor(minutes / 60);
	if (hours < 24) return `${hours}h ago`;
	const days = Math.floor(hours / 24);
	if (days < 7) return `${days}d ago`;
	return date.toLocaleDateString();
}
