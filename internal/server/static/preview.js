function previewImage(event, defaultImagePath) {
  event.preventDefault();
  const preview = document.getElementById("preview");

  if (event.target.files[0]) {
    const reader = new FileReader();
    reader.onload = (e) => {
      preview.src = e.target.result;
    };
    reader.readAsDataURL(event.target.files[0]);
    document.getElementById("save-image").classList.remove("hidden");
    document.getElementById("cancel").classList.remove("hidden");
  } else {
    preview.src = defaultImagePath;
  }
}

function editionCancel(evt, defaultImagePath) {
  evt.preventDefault();
  const preview = document.getElementById("preview");
  const fileInput = document.getElementById("image");
  preview.src = defaultImagePath;
  fileInput.value = "";
  document.getElementById("save-image").classList.add("hidden");
  document.getElementById("cancel").classList.add("hidden");
}

function showAndHide(formId) {
  const form = document.getElementById(formId);
  form.classList.toggle("hidden");
}

window.previewImage = previewImage;
window.editionCancel = editionCancel;
window.showAndHide = showAndHide;
