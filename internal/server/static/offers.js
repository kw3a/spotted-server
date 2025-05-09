function updateInputNames() {
  const problems = document.querySelectorAll("#problems-container .problem");
  problems.forEach((problem, pIndex) => {
    problem.querySelector("input[name^='problems'][name*='title']").name =
      `problems[${pIndex}][title]`;
    problem.querySelector("textarea[name^='problems'][name*='description']").name =
      `problems[${pIndex}][description]`;
    problem.querySelector("input[name^='problems'][name*='time_limit']").name =
      `problems[${pIndex}][time_limit]`;

    const testCases = problem.querySelectorAll(".tc");
    testCases.forEach((testCase, tcIndex) => {
      testCase.querySelector("textarea[name^='problems'][name*='test_cases'][name*='input']").name =
        `problems[${pIndex}][test_cases][${tcIndex}][input]`;
      testCase.querySelector("textarea[name^='problems'][name*='test_cases'][name*='output']").name =
        `problems[${pIndex}][test_cases][${tcIndex}][output]`;
    });

    const examples = problem.querySelectorAll(".ex");
    examples.forEach((example, exIndex) => {
      example.querySelector("textarea[name^='problems'][name*='examples'][name*='input']").name =
        `problems[${pIndex}][examples][${exIndex}][input]`;
      example.querySelector("textarea[name^='problems'][name*='examples'][name*='output']").name =
        `problems[${pIndex}][examples][${exIndex}][output]`;
    });
  });
}

function handleAddProblem(event) {
  event.preventDefault();
  const MAX_PROBLEMS = 5;
  const problemsContainer = document.getElementById("problems-container");
  const currentProblems = problemsContainer.children.length;
  if (currentProblems < MAX_PROBLEMS) {
    const newProblem = problemsContainer.firstElementChild.cloneNode(true);
    newProblem.querySelectorAll("input, textarea").forEach((field) => (field.value = ""));
    problemsContainer.appendChild(newProblem);
    updateInputNames();
  } else {
    showToast("No puedes agregar más de 5 problemas.");
  }
}

function handleDeleteProblem(event) {
  event.preventDefault();
  const MIN_PROBLEMS = 1;
  const problem = event.target.closest(".problem");
  const problemsContainer = document.getElementById("problems-container");
  const currentProblems = problemsContainer.children.length;
  if (currentProblems > MIN_PROBLEMS) {
    problem.remove();
    updateInputNames();
  } else {
    showToast("Debe haber al menos 1 problema.");
  }
}

function handleAddTestCase(event) {
  event.preventDefault();
  const MAX_ENTRIES = 10;
  const container = event.target.closest(".test-case-group");
  const entries = container.querySelectorAll(".test-case");
  if (entries.length < MAX_ENTRIES) {
    const newTestCase = entries[0].cloneNode(true);
    newTestCase.querySelectorAll("textarea").forEach((field) => (field.value = ""));
    container.appendChild(newTestCase);
    updateInputNames();
  } else {
    showToast("No puedes agregar más de 10 casos de prueba.");
  }
}

function handleDeleteTestCase(event) {
  event.preventDefault();
  const MIN_ENTRIES = 1;
  const container = event.target.closest(".test-case-group");
  const entries = container.querySelectorAll(".test-case");
  if (entries.length > MIN_ENTRIES) {
    const testCase = event.target.closest(".test-case");
    testCase.remove();
    updateInputNames();
  } else {
    showToast("Debe haber al menos 1 caso de prueba.");
  }
}

function handleAddExample(event) {
  event.preventDefault();
  const MAX_ENTRIES = 10;
  const container = event.target.closest(".example-group");
  const entries = container.querySelectorAll(".example");
  if (entries.length < MAX_ENTRIES) {
    const newExample = entries[0].cloneNode(true);
    newExample.querySelectorAll("textarea").forEach((field) => (field.value = ""));
    container.appendChild(newExample);
    updateInputNames();
  } else {
    showToast("No puedes agregar más de 10 ejemplos.");
  }
}

function handleDeleteExample(event) {
  event.preventDefault();
  const MIN_ENTRIES = 1;
  const container = event.target.closest(".example-group");
  const entries = container.querySelectorAll(".example");
  if (entries.length > MIN_ENTRIES) {
    const example = event.target.closest(".example");
    example.remove();
    updateInputNames();
  } else {
    showToast("Debe haber al menos 1 ejemplo.");
  }
}

const validateForm = () => {
  let valid = true;

  const validateField = (field, errorElement, min, max) => {
    const value = field.value.trim();
    if (value.length === 0) {
      errorElement.textContent = min ? `Este campo es obligatorio. (min. ${min} caracteres)` : "Este campo es obligatorio.";
      return false;
    } else if (min && value.length < min) {
      errorElement.textContent = `Este campo debe tener al menos ${min} caracteres.`;
      return false;
    } else if (max && value.length > max) {
      errorElement.textContent = `Este campo no puede tener más de ${max} caracteres.`;
      return false;
    } else {
      errorElement.textContent = "";
      return true;
    }
  };

  const offerTitle = document.getElementById("f-title");
  const titleError = document.getElementById("f-title-error");
  if (!validateField(offerTitle, titleError, 10, 60)) valid = false;

  const minWage = parseInt(document.getElementById("f-min-wage").value.trim(), 10);
  const maxWage = parseInt(document.getElementById("f-max-wage").value.trim(), 10);
  const wageError = document.getElementById("f-wage-error");
  if (isNaN(minWage) || isNaN(maxWage)) {
    wageError.textContent = "El rango salarial es obligatorio.";
    valid = false;
  } else if (minWage > maxWage) {
    wageError.textContent = "El salario mínimo no puede ser mayor.";
    valid = false;
  } else {
    wageError.textContent = "";
  }
  const about = document.getElementById("f-about");
  const aboutError = document.getElementById("f-about-error");
  if (!validateField(about, aboutError, 200, 5000)) valid = false;

  const requirements = document.getElementById("f-requirements");
  const requirementsError = document.getElementById("f-requirements-error");
  if (!validateField(requirements, requirementsError, 200, 5000)) valid = false;

  const benefits = document.getElementById("f-benefits");
  const benefitsError = document.getElementById("f-benefits-error");
  if (!validateField(benefits, benefitsError, 200, 5000)) valid = false;

  const validateProblems = () => {
    let allValid = true;
    const problems = Array.from(document.querySelectorAll("#problems-container .problem"));

    problems.forEach((problem) => {
      const title = problem.querySelector("input[name^='problems'][name*='title']");
      const titleError = title.nextElementSibling;
      const isTitleValid = validateField(title, titleError, 1, 64);
      if (!isTitleValid) allValid = false;

      const description = problem.querySelector("textarea[name^='problems'][name*='description']");
      const descriptionError = description.nextElementSibling;
      const isDescriptionValid = validateField(description, descriptionError, 10, 5000);
      if (!isDescriptionValid) allValid = false;

      const validateGroup = (groupSelector) => {
        const items = Array.from(problem.querySelectorAll(groupSelector));
        return items.forEach((item) => {
          const input = item.querySelector("textarea[name*='input']");
          const output = item.querySelector("textarea[name*='output']");
          const inputError = input.nextElementSibling;
          const outputError = output.nextElementSibling;
          const isInputValid = validateField(input, inputError, 1, 1000);
          const isOutputValid = validateField(output, outputError, 1, 1000);
          if (!isInputValid || !isOutputValid) allValid = false;
          return isInputValid && isOutputValid;
        });
      };

      validateGroup(".test-case-group .test-case");
      validateGroup(".example-group .example");
    });

    return allValid;
  };

  if (!validateProblems()) valid = false;

  const languages = Array.from(document.querySelectorAll("input[name='quiz[languages]']:checked"));
  const languagesError = document.getElementById("f-languages-error");
  if (languages.length === 0) {
    languagesError.textContent = "Selecciona al menos un lenguaje.";
    valid = false;
  } else {
    languagesError.textContent = "";
  }

  return valid;
};

function submitForm(evt) {
  evt.preventDefault();
  const valid = validateForm();
  if (valid) {
    htmx.trigger('#offer-form-container', "evtsubmitoffer");
  } else {
    showToast("Por favor, llena correctamente todos los campos.");
  }
}

function toggleDetails(event) {
  const detailsElement = event.target.closest('details');
  if (detailsElement.hasAttribute('open')) {
    detailsElement.removeAttribute('open');
  } else {
    detailsElement.setAttribute('open', true);
  }
}
function errorImage() {
  const img = document.createElement("img");
  img.width = 20;
  img.height = 20;
  img.src = "/public/cancel.svg";
  img.alt = "Error";
  return img;
}

function showToast(message) {
  const toastContainer = document.getElementById("toast");

  // Create the toast element
  const toast = document.createElement("div");
  toast.className =
    "flex items-center p-4 space-x-4 divide-x rounded-lg shadow-sm text-red-500 divide-gray-700 bg-gray-800";

  const img = errorImage();
  toast.appendChild(img);

  const text = document.createElement("p");
  text.className = "ps-4 text-sm font-normal";
  text.textContent = message;
  toast.appendChild(text);

  // Add the toast to the container
  toastContainer.appendChild(toast);

  // Make the toast visible
  setTimeout(() => {
    toast.classList.remove("opacity-0");
  }, 100);

  // Remove the toast after 3 seconds
  setTimeout(() => {
    toast.classList.add("opacity-0");
    setTimeout(() => {
      toast.remove();
    }, 300); // Matches the transition duration
  }, 5000);
}
