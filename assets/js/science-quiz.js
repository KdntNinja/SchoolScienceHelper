// Science Quiz UI logic (fetches questions, renders, handles answers)
(async function() {
  const root = document.getElementById('science-quiz-root');
  if (!root) return;
  let subject = 'Biology'; // Default, can add selector later
  let topic = '';
  let questions = [];
  let current = 0;
  let correct = 0;
  let progress = {};

  const subjectSelect = document.getElementById('subject-select');
  const topicSelect = document.getElementById('topic-select');
  const startBtn = document.getElementById('start-quiz-btn');

  // Example topics per subject (replace/expand with real topics as needed)
  const topicsBySubject = {
    Biology: ["Cell Biology", "Organisation", "Infection & Response", "Bioenergetics"],
    Physics: ["Forces", "Energy", "Waves", "Electricity"],
    Chemistry: ["Atomic Structure", "Bonding", "Quantitative Chemistry", "Chemical Changes"]
  };

  function populateTopics(subject) {
    topicSelect.innerHTML = '<option value="">All Topics</option>';
    (topicsBySubject[subject] || []).forEach(t => {
      const opt = document.createElement('option');
      opt.value = t;
      opt.textContent = t;
      topicSelect.appendChild(opt);
    });
  }

  async function loadQuestions() {
    const res = await fetch(`/api/science/questions?subject=${encodeURIComponent(subject)}&topic=${encodeURIComponent(topic)}`);
    questions = await res.json();
    current = 0;
    correct = 0;
    render();
  }

  async function submitAnswer(idx) {
    const q = questions[current];
    const res = await fetch('/api/science/answer', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ question_id: q.id, selected: idx })
    });
    const data = await res.json();
    if (data.correct) correct++;
    progress[q.id] = data;
    current++;
    render();
  }

  function render() {
    if (!questions.length) {
      root.innerHTML = '<div class="text-center text-lg">No questions found for this subject/topic.</div>';
      return;
    }
    if (current >= questions.length) {
      root.innerHTML = `<div class='text-center'><h2 class='text-2xl font-bold mb-4 text-primary'>Quiz Complete!</h2><div class='mb-2'>Score: ${correct} / ${questions.length}</div><button class='btn bg-primary text-white px-4 py-2 rounded' onclick='window.location.reload()'>Try Again</button></div>`;
      return;
    }
    const q = questions[current];
    let html = `<div class='mb-4 text-lg font-semibold'>Q${current+1}: ${q.question}</div>`;
    html += `<ul class='space-y-2'>`;
    q.choices.forEach((c,i)=>{
      html += `<li><button class='btn w-full bg-primary/90 hover:bg-primary text-white font-bold px-4 py-2 rounded' onclick='window.submitScienceAnswer(${i})'>${c}</button></li>`;
    });
    html += `</ul>`;
    html += `<div class='mt-6 text-sm text-muted-foreground'>Question ${current+1} of ${questions.length}</div>`;
    root.innerHTML = html;
  }

  window.submitScienceAnswer = submitAnswer;
  if (subjectSelect && topicSelect && startBtn) {
    subjectSelect.addEventListener('change', e => {
      subject = subjectSelect.value;
      populateTopics(subject);
    });
    startBtn.addEventListener('click', () => {
      subject = subjectSelect.value;
      topic = topicSelect.value;
      loadQuestions();
    });
    populateTopics(subjectSelect.value);
  }
  await loadQuestions();
})();
