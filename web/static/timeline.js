const timeline = document.querySelector('.timeline');
if (timeline && 'IntersectionObserver' in window) {
  const observer = new IntersectionObserver((entries) => entries.forEach((entry) => entry.target.classList.toggle('is-current', entry.isIntersecting)), {root: timeline, threshold: .65});
  timeline.querySelectorAll('.timeline-card').forEach((card) => observer.observe(card));
}
