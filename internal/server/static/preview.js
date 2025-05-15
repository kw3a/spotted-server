const $ = (el) => document.querySelector(el);

function showAndHide(...IDs) {
  IDs.forEach((id) => {
    const tag = document.getElementById(id);
    tag?.classList.toggle("hidden");
  });
}

function triggers() {
  const events = [
    { event: "link-added", selector: "#empty-links", form: "#linkForm"},
    { event: "skill-added", selector: "#empty-skills", form: "#skillForm"},
    { event: "ed-added", selector: "#empty-ed", form: "#edForm" },
    { event: "exp-added", selector: "#empty-exp", form: "#expForm" },
  ];
  events.forEach(({ event, selector, form }) => {
    document.body.addEventListener(event, () => {
      $(selector)?.classList.add("hidden");
      $(form)?.close();
    });
  });
}

document.addEventListener("DOMContentLoaded", () => {
  const preview = $("#preview");
  if (preview) {
    let imageUrl = preview.src;
    let currentUserImgPath = imageUrl || "/public/user.svg";

    const previewInput = $("#preview-input");
    if (previewInput) {
      previewInput.onchange = (event) => {
        event.preventDefault();
        const preview = $("#preview");
        if (event.target.files[0]) {
          const reader = new FileReader();
          reader.onload = (e) => { preview.src = e.target.result; };
          reader.readAsDataURL(event.target.files[0]);
          $("#edit-options")?.classList.remove("hidden");
        } else {
          preview.src = currentUserImgPath;
          previewInput.value = "";
          $("#edit-options")?.classList.add("hidden");
        }
      };
    }

    const cancel = $("#cancel");
    if (cancel) {
      cancel.onclick = (evt) => {
        evt.preventDefault();
        $("#preview").src = currentUserImgPath;
        if (previewInput) previewInput.value = "";
        $("#edit-options")?.classList.add("hidden");
      };
    }

    document.body.addEventListener("image-changed", () => {
      currentUserImgPath = $("#preview")?.src;
      $("#edit-options")?.classList.add("hidden");
    });
  }
  triggers();
});

window.showAndHide = showAndHide;
