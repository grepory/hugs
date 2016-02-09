INSERT INTO notification_types VALUES ('1', 'slack_bot');
INSERT INTO notification_types VALUES ('2', 'slack_hook');
INSERT INTO notification_types VALUES ('3', 'email');

INSERT INTO notifications VALUES ('1','6666', uuid_in('5963d7bc-6ba2-11e5-8603-6ba085b2f5b5'), '13', 'email', 'dan@opsee.co');
INSERT INTO notifications VALUES ('2','6666', uuid_in('5963d7bc-6ba2-11e5-8603-6ba085b2f5b5'), '13', 'slack_hook', 'https://hooks.slack.com/services/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');

