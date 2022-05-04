import github from '../src/github.js';
import slack from '../src/slack.js';

const count = await github();
slack(count);
