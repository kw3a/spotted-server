
function previewImage(event, defaultPath) {
  event.preventDefault();
  const preview = document.getElementById("preview");
  if (!preview) return;
  if (event.target.files[0]) {
    const reader = new FileReader();
    reader.onload = (e) => {
      preview.src = e.target.result;
    };
    reader.readAsDataURL(event.target.files[0]);
    document.getElementById("edit-options")?.classList.remove("hidden");
  } else {
    preview.src = defaultPath;
    event.target.value = "";
    document.getElementById("edit-options")?.classList.add("hidden");
  }
}

window.previewImage = previewImage;