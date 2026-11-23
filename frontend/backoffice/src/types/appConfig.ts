// GET / PUT /app-configs/email-contact-card — the contact card image at the bottom of emails.
export type EmailContactCard = {
  imageUrl: string; // uploaded card; '' = the built-in one
  defaultImage: string; // the built-in card as a data: URL ('' in the mock)
  updateDate: string | null; // null until an admin saves one
  updateBy: string; // name of who saved it
};

export type EmailContactCardUpdateRequest = { imageUrl: string }; // '' = back to the built-in card
