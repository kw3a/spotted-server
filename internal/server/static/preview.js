function previewImage(event, defaultImagePath) {
  event.preventDefault();
  const preview = document.getElementById("preview");

  if (event.target.files[0]) {
    const reader = new FileReader();
    reader.onload = (e) => {
      preview.src = e.target.result;
    };
    reader.readAsDataURL(event.target.files[0]);
    document.getElementById("edit-options").classList.remove("hidden");
  } else {
    preview.src = defaultImagePath;
    const fileInput = document.getElementById("image");
    fileInput.value = "";
    document.getElementById("edit-options").classList.add("hidden");
  }
}

function editionCancel(evt, defaultImagePath) {
  evt.preventDefault();
  const preview = document.getElementById("preview");
  const fileInput = document.getElementById("image");
  preview.src = defaultImagePath;
  fileInput.value = "";
  document.getElementById("edit-options").classList.add("hidden");
}

function showAndHide(...IDs) {
  IDs.forEach(id => {
    const form = document.getElementById(id);
    form?.classList.toggle("hidden");
  });
}

function toggleEditable(...IDs) {
  IDs.forEach(id => {
    const form = document.getElementById(id);
    form?.toggleAttribute("contenteditable");
  });
}

function getText(ID) {
  const tag = document.getElementById(ID);
  if (tag) {
    return tag.textContent.trim();
  }
  return ""
}

let savedPDesc = getText('pDesc')
function resetPDesc(ID) {
  const tag = document.getElementById(ID);
  if (tag) {
    tag.textContent = savedPDesc
  }
}

function updatePDesc(ID) {
  const tag = document.getElementById(ID);
  if (tag) {
    savedPDesc = tag.textContent.trim();
  }
}
window.previewImage = previewImage;
window.editionCancel = editionCancel;
window.showAndHide = showAndHide;
window.toggleEditable = toggleEditable;
window.getText = getText;
window.resetPDesc = resetPDesc;
window.updatePDesc = updatePDesc;
