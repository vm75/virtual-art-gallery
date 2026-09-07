const dialog = document.querySelector('#lightbox');
if (dialog) {
  const image = dialog.querySelector('img');
  const detail = dialog.querySelector('.dialog-detail');
  const links = [...document.querySelectorAll('[data-lightbox]')];
  let current = 0;
  let lastFocus;
  const show = (index) => { current = (index + links.length) % links.length; const link = links[current]; image.src = link.querySelector('img').currentSrc || link.querySelector('img').src; image.alt = link.querySelector('img').alt; detail.href = link.href; };
  links.forEach((link, index) => link.addEventListener('click', (event) => {
    event.preventDefault(); lastFocus = document.activeElement;
    show(index); dialog.showModal();
  }));
  dialog.addEventListener('close', () => lastFocus?.focus());
  dialog.querySelector('.dialog-close').addEventListener('click', () => dialog.close());
  dialog.querySelector('.dialog-prev').addEventListener('click', () => show(current - 1));
  dialog.querySelector('.dialog-next').addEventListener('click', () => show(current + 1));
  dialog.addEventListener('keydown', (event) => { if (event.key === 'ArrowLeft') show(current - 1); if (event.key === 'ArrowRight') show(current + 1); });
  let startX = 0; dialog.addEventListener('touchstart', (event) => { startX = event.changedTouches[0].clientX; }, {passive: true}); dialog.addEventListener('touchend', (event) => { const delta = event.changedTouches[0].clientX - startX; if (Math.abs(delta) > 40) show(current + (delta < 0 ? 1 : -1)); }, {passive: true});
}
