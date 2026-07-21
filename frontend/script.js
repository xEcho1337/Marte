const drawer = document.getElementById('attentionDrawer');
const toggleDrawer = document.getElementById('toggleDrawer');
const closeDrawer = document.getElementById('closeDrawer');
const sidebarButtons = document.querySelectorAll('[data-sidebar-btn]');

function setDrawer(open) {
    drawer.classList.toggle('open', open);
    drawer.setAttribute('aria-hidden', String(!open));
    toggleDrawer.setAttribute('aria-expanded', String(open));
}

toggleDrawer.addEventListener('click', () => {
    setDrawer(!drawer.classList.contains('open'));
});

closeDrawer.addEventListener('click', () => setDrawer(false));

document.addEventListener('keydown', (event) => {
    if (event.key === 'Escape') {
        setDrawer(false);
    }
});

sidebarButtons.forEach((button) => {
    button.addEventListener('click', () => {
        sidebarButtons.forEach((other) => {
            other.classList.remove('btn-outline-primary');
            other.classList.add('text-start', 'w-100');
        });

        button.classList.add('btn-outline-primary');
    });
});
