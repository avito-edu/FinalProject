INSERT INTO questions (text, options, correct_id, is_active)
VALUES
(
  'Столица Франции?',
  '[{"id":"A","text":"Berlin"},{"id":"B","text":"Paris"},{"id":"C","text":"Madrid"},{"id":"D","text":"Rome"}]',
  'B',
  true
),
(
  'Самая большая планета Солнечной системы?',
  '[{"id":"A","text":"Mars"},{"id":"B","text":"Earth"},{"id":"C","text":"Jupiter"},{"id":"D","text":"Venus"}]',
  'C',
  true
),
(
  'Сколько будет 2 + 2?',
  '[{"id":"A","text":"3"},{"id":"B","text":"4"},{"id":"C","text":"5"},{"id":"D","text":"22"}]',
  'B',
  true
)
ON CONFLICT (text) DO NOTHING;
