import { test, expect } from './fixtures';

test.describe('Markdown links', () => {
  test('external links have target="_blank" and rel="noopener noreferrer"', async ({
    beans,
    backlogPage,
    page
  }) => {
    const id = beans.create('Link Test Bean', {
      status: 'todo',
      type: 'task'
    });

    // Update the bean body with a markdown link and a bare URL
    beans.run([
      'update',
      id,
      '--body-append',
      'Visit [Example](https://example.com) or https://example.org for more info.'
    ]);

    await backlogPage.goto(1);
    await backlogPage.selectBean('Link Test Bean');

    // Wait for the bean detail body to render
    const body = page.locator('.bean-body');
    await expect(body).toBeVisible({ timeout: 10_000 });

    // Check the explicit markdown link
    const markdownLink = body.locator('a[href="https://example.com"]');
    await expect(markdownLink).toHaveAttribute('target', '_blank');
    await expect(markdownLink).toHaveAttribute('rel', 'noopener noreferrer');

    // Check the autolinked bare URL
    const autoLink = body.locator('a[href="https://example.org"]');
    await expect(autoLink).toHaveAttribute('target', '_blank');
    await expect(autoLink).toHaveAttribute('rel', 'noopener noreferrer');
  });
});
