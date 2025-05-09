import hljs from 'https://cdnjs.cloudflare.com/ajax/libs/highlight.js/11.9.0/es/highlight.min.js';

function openModal(id) {
  document.querySelectorAll('[id^="modal-"]').forEach((modal) => {
    modal.classList.add("hidden");
  });
  document.getElementById(id).classList.remove("hidden");
}
window.openModal = openModal

hljs.highlightAll();
