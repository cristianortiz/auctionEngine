-- Delete tables in reverse order to avoid FK dependency issuesDROP TABLE IF EXISTS bids;
DROP TABLE IF EXISTS auction_lots;

-- delete the trigger and the function associated if  exists
DROP TRIGGER IF EXISTS update_auction_lots_updated_at ON auction_lots;
DROP FUNCTION IF EXISTS update_updated_at_column();